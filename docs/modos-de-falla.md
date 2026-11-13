# Modos de falla del servicio de productos

Este documento sigue la regla del DAS: antes de programar algo que se prueba
aislado, se escriben las fallas posibles, qué las impide y cómo se verifican.
En el Hito 1 los endpoints usan datos simulados; las validaciones marcadas como
"DB" se implementan en el Hito 2.

## Resumen

| Caso | Capa que lo previene | Respuesta esperada | Verificación |
|---|---|---|---|
| SKU duplicado concurrente | Service + `UNIQUE(sku)` en PostgreSQL | `409 Conflict` | Dos `POST` simultáneos con el mismo SKU: uno `201`, uno `409` |
| `categoria_ids` con UUID inexistente | Service + FK de la tabla puente | `422 Unprocessable Entity` | `POST` con un UUID válido pero inexistente |
| `categoria_ids` vacío | Contrato Huma (`minItems:"1"`) | `422 Unprocessable Entity` | `POST` con `categoria_ids: []` |
| `categoria_ids` con duplicados | Contrato Huma (`uniqueItems:"true"`) + PK de la tabla puente | `422 Unprocessable Entity` | `POST` con el mismo UUID dos veces |
| Costo negativo | Contrato Huma (`minimum:"0"`) + `CHECK (costo >= 0)` | `422 Unprocessable Entity` | `POST` con `costo: -1` |
| Vender con `permite_venta=false` | Service del flujo de venta | `409 Conflict` o `422` | Se verifica cuando exista el endpoint de venta o en el servicio consumidor |

## 1. SKU duplicado concurrente

**Por qué pasa.** Dos requests simultáneas leen que el SKU no existe y ambas
intentan insertar. Una comprobación previa en el service no basta.

**Cómo se previene.**
- El service puede hacer una consulta previa para dar un error temprano, pero
  la garantía real es la restricción `UNIQUE(sku)` de PostgreSQL.
- El store captura la violación de unicidad (SQLSTATE `23505`) y la traduce a
  un error de dominio.
- El handler traduce ese error a `409 Conflict`; el formato del cuerpo es
  `application/problem+json` de Huma (RFC 9457).

**Cómo se verifica en e2e.**
- Lanzar dos `POST /api/productos` concurrentes con el mismo SKU.
- El reporte determinista debe mostrar exactamente un `201` y un `409`.
- Romper a propósito la constraint `UNIQUE(sku)` y confirmar que el e2e falla
  (dos `201`). Ese es el control del DAS §9.5.

## 2. `categoria_ids` con un UUID inexistente

**Por qué pasa.** `format:"uuid"` solo valida la sintaxis. Un UUID bien formado
puede no existir en el catálogo de categorías.

**Cómo se previene.**
- En el contrato, `categoria_ids` es un array de UUIDs con `minItems: 1` y
  `uniqueItems: true`.
- En el service se verifica que cada `categoria_id` exista. La fuente puede ser
  una tabla local de categorías o la API del servicio dueño de ese dato.
- En la base de datos, la tabla puente
  `producto_categorias(producto_id, categoria_id)` tiene FK contra categorías.

**Cómo se verifica en e2e.**
- `POST /api/productos` con un UUID aleatorio bien formado que no exista.
- Esperar `422` con detalle del campo `body.categoria_ids`.
- Si la categoría se valida contra otro servicio, agregar un caso con ese
  servicio caído y verificar que la escritura no se confirme.

## 3. `categoria_ids` vacío

**Por qué pasa.** El cliente envía `"categoria_ids": []` o no envía el campo.

**Cómo se previene.**
- El contrato lo rechaza antes de llegar al handler: `required:"true"` y
  `minItems:"1"`.
- No se agrega una regla equivalente en el service porque el contrato ya la
  cubre; el service solo debería recibir arrays no vacíos.

**Cómo se verifica en e2e.**
- `POST /api/productos` con `"categoria_ids": []` → `422`.
- `POST /api/productos` sin `categoria_ids` → `422`.

## 4. `categoria_ids` con duplicados

**Por qué pasa.** El cliente repite el mismo UUID dentro del array.

**Cómo se previene.**
- El contrato usa `uniqueItems:"true"`; Huma rechaza el cuerpo antes del
  handler.
- En la base de datos, la PK compuesta
  `(producto_id, categoria_id)` impide duplicados aunque la validación de
  contrato se desactive.

**Cómo se verifica en e2e.**
- `POST /api/productos` con el mismo UUID dos veces → `422`.
- En el Hito 2, intentar insertar dos veces la misma relación directamente en
  la base y verificar que la PK falla.

## 5. Costo negativo

**Por qué pasa.** El cliente envía un número negativo o un decimal.

**Cómo se previene.**
- El contrato usa `int64` y `minimum:"0"`. Huma rechaza negativos y valores no
  enteros.
- En la base de datos, `costo` es `bigint` con `CHECK (costo >= 0)`.
- Nunca se usa `float` para montos.

**Cómo se verifica en e2e.**
- `POST /api/productos` con `"costo": -1` → `422`.
- `POST /api/productos` con `"costo": 12.5` → `422`.
- Romper el `CHECK` en la base y confirmar que el e2e de persistencia falla.

## 6. Intento de vender un producto con `permite_venta=false`

**Por qué pasa.** Un flujo de venta intenta usar un producto que no está
habilitado para venderse.

**Cómo se previene.**
- Este servicio expone `permite_venta` como dato autoritativo.
- La regla pertenece al service del flujo de venta (o al servicio consumidor
  que venda el producto): antes de confirmar la venta, consulta el producto y
  rechaza si `permite_venta` es `false`.
- No se implementa en el handler de productos porque el handler solo traduce
  HTTP.

**Cómo se verifica en e2e.**
- Cuando exista el endpoint de venta: intentar vender un producto con
  `permite_venta=false` y esperar `409` o `422`, sin cambios de stock ni de
  estado.
- Mientras no exista ese endpoint, la regla queda documentada y el contrato
  mantiene el flag visible.

## Modelo de categorías múltiples (Opción A)

La decisión de modelado y las preguntas abiertas para el equipo están en
[`decisiones.md`](./decisiones.md). Para el Hito 2 se anota la tabla puente,
sin implementarla todavía:

```sql
-- Hito 2: migración goose, no forma parte del Hito 1.
CREATE TABLE producto_categorias (
    producto_id  uuid NOT NULL REFERENCES productos(id) ON DELETE CASCADE,
    categoria_id uuid NOT NULL REFERENCES categorias(id),
    PRIMARY KEY (producto_id, categoria_id)
);
```
