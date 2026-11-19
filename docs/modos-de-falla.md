# Modos de falla del servicio de productos

Este documento sigue la regla del DAS: antes de programar algo que se prueba
aislado, se escriben las fallas posibles, qué las impide y cómo se verifican.
En el Hito 1 los endpoints usan datos simulados; lo marcado como "DB" se
implementa en el Hito 2.

## Resumen

| Caso | Capa que lo previene | Respuesta esperada | Verificación |
|---|---|---|---|
| Secuencial de SKU duplicado (concurrente) | Service + secuencia en la misma transacción + `UNIQUE(sku)` | Reintento interno; si se agota, `409` | Dos `POST` simultáneos al mismo departamento: SKU distintos |
| `departamento_id` inexistente | Service + FK de `productos` | `422` | `POST` con un UUID válido pero inexistente |
| `categoria_ids` con UUID inexistente | Service + FK de la tabla puente | `422` | `POST` con un UUID válido pero inexistente |
| `categoria_ids` vacío | Contrato Huma (`minItems:"1"`) | `422` | `POST` con `categoria_ids: []` |
| `categoria_ids` con duplicados | Contrato Huma (`uniqueItems:"true"`) + PK de la tabla puente | `422` | `POST` con el mismo UUID dos veces |
| `costo_unitario` negativo o decimal | Contrato Huma (`int64`, `minimum:"0"`) + `CHECK` | `422` | `POST` con `-1` y con `1.5` |
| `tipo_impuesto` inválido | Contrato Huma (`enum`) + `CHECK` | `422` | `POST` con `"OTRO"` |
| Intento de fijar o cambiar el SKU | Contrato Huma (`sku` fuera de los cuerpos) | `422` `unexpected property` | `POST` y `PATCH` con `sku` |
| Vender con `permite_venta=false` | Service del flujo de venta | `409` o `422` | Cuando exista el endpoint de venta |

## 1. Secuencial de SKU duplicado (concurrente)

**Por qué pasa.** El SKU ya no lo elige el cliente: lo genera el servicio con un
secuencial por departamento. Dos creaciones simultáneas en el mismo departamento
pueden leer el mismo contador y producir el mismo SKU.

**Cómo se previene.**
- En el Hito 1 el handler simulado serializa con un mutex; eso sirve para el mock,
  no es una garantía real.
- En el Hito 2 el secuencial se obtiene dentro de la **misma transacción** que
  inserta el producto (secuencia o `UPDATE ... RETURNING` sobre el contador del
  departamento).
- `UNIQUE(sku)` es la garantía final: si dos transacciones calculan el mismo
  valor, una falla con SQLSTATE `23505`.
- El store traduce esa violación a un error de dominio y el service reintenta con
  el siguiente secuencial un número acotado de veces; si se agota, el handler
  responde `409`.

**Cómo se verifica en e2e.**
- Lanzar N `POST /api/productos` concurrentes al mismo departamento con nombres
  distintos.
- El reporte determinista debe mostrar N productos creados y N SKUs distintos,
  todos con el mismo prefijo y secuenciales contiguos.
- Romper a propósito el `UNIQUE(sku)` y confirmar que el e2e falla.

## 2. `departamento_id` inexistente

**Por qué pasa.** `format: uuid` solo valida la sintaxis; el departamento puede no
existir.

**Cómo se previene.**
- El service verifica que el departamento exista antes de generar el SKU.
- En la base, `productos.departamento_id` tiene FK contra `departamentos`.

**Cómo se verifica en e2e.**
- `POST` con un UUID bien formado que no exista → `422` con detalle en
  `body.departamento_id`.
- Confirmar que no se consumió ningún secuencial del departamento.

## 3. `categoria_ids` con un UUID inexistente

**Por qué pasa.** Igual que el caso anterior: el formato es válido, el dato no
existe.

**Cómo se previene.**
- El service verifica cada `categoria_id` contra la fuente de categorías.
- La tabla puente `producto_categorias(producto_id, categoria_id)` tiene FK.

**Cómo se verifica en e2e.**
- `POST` con un UUID aleatorio bien formado → `422` con detalle en
  `body.categoria_ids`.
- Si la validación es contra otro servicio, agregar el caso con ese servicio
  caído y verificar que la escritura no se confirme.

## 4. `categoria_ids` vacío

**Por qué pasa.** El cliente envía `[]` o no envía el campo.

**Cómo se previene.** Contrato: `required:"true"` y `minItems:"1"`. Huma rechaza
antes del handler, así que el service solo recibe arrays no vacíos.

**Cómo se verifica en e2e.** `POST` con `[]` → `422`; `POST` sin el campo → `422`.

## 5. `categoria_ids` con duplicados

**Por qué pasa.** El cliente repite el mismo UUID.

**Cómo se previene.** `uniqueItems:"true"` en el contrato, y en la base la PK
compuesta de la tabla puente.

**Cómo se verifica en e2e.** `POST` con el mismo UUID dos veces → `422`; en el
Hito 2, insertar dos veces la relación directo en la base y confirmar que la PK
falla.

## 6. `costo_unitario` negativo o decimal

**Por qué pasa.** El cliente envía un número negativo o con decimales.

**Cómo se previene.** Contrato: `int64` y `minimum:"0"`. Un decimal no entra en
`int64`, así que Huma responde `422`. En la base,
`CHECK (costo_unitario >= 0)`.

**Cómo se verifica en e2e.** `POST` con `-1` → `422`; `POST` con `1.5` → `422`;
romper el `CHECK` y confirmar que el e2e de persistencia falla.

## 7. `tipo_impuesto` inválido

**Por qué pasa.** El cliente envía un valor fuera del catálogo (`OTRO`,
`iva`, vacío).

**Cómo se previene.** El contrato declara `enum: [IVA, EXENTO]`; Huma responde
`422`. En la base, `CHECK (tipo_impuesto IN ('IVA','EXENTO'))`.

**Cómo se verifica en e2e.** `POST` con `"OTRO"` → `422`; omitir el campo → `422`.

## 8. Intento de fijar o cambiar el SKU

**Por qué pasa.** Un cliente acostumbrado al modelo anterior envía `sku` en el
`POST` o intenta cambiarlo por `PATCH`.

**Cómo se previene.**
- `sku` **no existe** en `CrearProductoBody` ni en `ActualizarProductoBody`, y
  `departamento_id` tampoco está en el `PATCH`: cambiarlo cambiaría el prefijo del
  SKU.
- Huma rechaza campos desconocidos: responde `422` con `"unexpected property"`
  indicando `body.sku`.
- En el Hito 2 el `UPDATE` no incluye la columna `sku`.

**Cómo se verifica en e2e.**
- `POST /api/productos` con `sku` en el cuerpo → `422` con `body.sku`.
- `PATCH /api/productos/{id}` con `sku` o con `departamento_id` → `422`.
- `GET` posterior: el SKU y el departamento originales no cambiaron.

## 9. Intento de vender un producto con `permite_venta=false`

**Por qué pasa.** Un flujo de venta intenta usar un producto no habilitado.

**Cómo se previene.** Este servicio expone `permite_venta` como dato autoritativo;
la regla pertenece al service del flujo de venta, que consulta el producto antes
de confirmar.

**Cómo se verifica en e2e.** Cuando exista el endpoint de venta: intentar vender
un producto con `permite_venta=false` y esperar `409` o `422`, sin cambios de
estado.

## Esquema previsto para el Hito 2

No implementado. Se documenta para que el contrato y el esquema no se
contradigan.

```sql
-- Hito 2: migración goose, no forma parte del Hito 1.
CREATE TABLE departamentos (
    id     uuid PRIMARY KEY,
    codigo text NOT NULL UNIQUE,
    nombre text NOT NULL
);

CREATE TABLE productos (
    id             uuid PRIMARY KEY,
    sku            text NOT NULL UNIQUE,
    departamento_id uuid NOT NULL REFERENCES departamentos(id),
    nombre         text NOT NULL,
    descripcion    text NOT NULL DEFAULT '',
    costo_unitario bigint NOT NULL CHECK (costo_unitario >= 0),
    tipo_impuesto  text NOT NULL CHECK (tipo_impuesto IN ('IVA','EXENTO')),
    permite_venta  boolean NOT NULL,
    activo         boolean NOT NULL DEFAULT true,
    creado_en      timestamptz NOT NULL DEFAULT now(),
    actualizado_en timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE producto_categorias (
    producto_id  uuid NOT NULL REFERENCES productos(id) ON DELETE CASCADE,
    categoria_id uuid NOT NULL,
    PRIMARY KEY (producto_id, categoria_id)
);
```
