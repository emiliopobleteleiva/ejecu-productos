# Decisiones de modelado

## Estado

Decisiones adoptadas para el **contrato v1** e implementadas en el Hito 1. La
persistencia queda para el Hito 2. Cada punto indica qué se decidió, dónde vive
en el contrato y qué implica en la base de datos.

> Estas decisiones se tomaron antes de habilitar `oasdiff` en CI, que es la
> última ventana para cambiar el contrato sin publicar `/api/v2`.

## 1. SKU autogenerado e inmutable por departamento

**Decisión.** El cliente no envía el SKU. El servicio lo genera con el formato
`<CODIGO_DEPARTAMENTO>-<SECUENCIAL>` (seis dígitos), por ejemplo `BEB-000001`. El
secuencial es propio de cada departamento y el SKU no se modifica nunca.

| Dónde | Efecto |
|---|---|
| `Producto.sku` | `readOnly: true`: se publica en la respuesta |
| `CrearProductoBody` | **no existe** `sku` |
| `ActualizarProductoBody` | **no existe** `sku` ni `departamento_id` |

**Hito 2.** Tabla `departamentos(codigo)`, secuencia por departamento generada
dentro de la misma transacción del `INSERT` y `UNIQUE(sku)` como garantía final.

**Consecuencia.** Un cliente no puede imponer un SKU existente. Importar un
catálogo desde otro sistema requiere un endpoint o proceso aparte, no este
contrato.

## 2. Departamento

**Decisión.** Todo producto pertenece a un departamento. `departamento_id` es un
UUID obligatorio al crear, inmutable después, y filtrable en el listado.

**Hito 2.** FK a la tabla `departamentos`. El código que usa el SKU es un dato del
departamento, no se deriva del UUID: en el Hito 1 hay un catálogo simulado en el
handler.

**Abierto.** Si los departamentos viven en otro servicio, hay que decidir si se
validan por API en cada escritura o si se mantiene una copia local.

## 3. Costo unitario con IVA

**Decisión.** El campo se llama `costo_unitario`, es `int64` entero en pesos
chilenos e **incluye IVA**. El servicio no calcula impuestos: informa el valor
final que el negocio define.

**Contrato.** `minimum: 0`. Nunca `float`, ni en el contrato ni en la base.

**Hito 2.** `costo_unitario bigint NOT NULL CHECK (costo_unitario >= 0)`.

## 4. Tipo de impuesto

**Decisión.** `tipo_impuesto` es obligatorio al crear y solo admite `IVA` o
`EXENTO`. No se deduce del costo: un producto puede tener costo y estar exento.

**Contrato.** `enum: [IVA, EXENTO]`, validado por Huma antes del handler.

**Hito 2.** `CHECK (tipo_impuesto IN ('IVA','EXENTO'))` o un enum de PostgreSQL.

## 5. Identificadores

**Decisión.** Todos los identificadores expuestos en la API son **UUID**
(`format: uuid`): `id` del producto, `departamento_id` y cada `categoria_ids`. El
**identificador de negocio es el SKU**, autogenerado e inmutable.

**Motivo.** Los UUID no filtran volumen, no colisionan entre servicios y permiten
componer las APIs en el ensamblaje LPS. El SKU es para humanos y para el negocio,
no para rutas ni claves foráneas.

## 6. Categorización múltiple como lista simple (Opción A)

**Decisión.** `categoria_ids` es un array de UUIDs, obligatorio, con `minItems: 1`
y `uniqueItems: true`. "Lista simple" se interpreta como **array de ids**, no como
objetos anidados con nombre y descripción.

| Opción descartada | Por qué |
|---|---|
| `categoria_id` singular | Migrar después a múltiples rompe el contrato |
| Objetos de categoría anidados | Acopla el contrato a datos de otro servicio |

**Hito 2.** Tabla puente `producto_categorias(producto_id, categoria_id)` con PK
compuesta y FKs.

**Abierto.** Si el equipo quiso decir "lista de nombres o códigos" en vez de ids,
es un cambio de contrato y hay que decidirlo antes de habilitar `oasdiff`.

## 7. Una instancia por cliente

**Decisión.** El despliegue es **single-tenant**: cada cliente recibe su propia
instancia con su propia base PostgreSQL. Por eso el contrato **no** incluye
`cliente_id` ni filtros por cliente.

**Si cambia a multi-tenant** hace falta `cliente_id` en todas las tablas, filtros
derivados del JWT y una revisión de permisos: es un contrato nuevo, no un campo
más.

## Preguntas abiertas para el equipo

1. ¿La categoría es un dato propio de este servicio o vive en otro? Si vive en
   otro, ¿se valida por API en cada escritura o con copia local?
2. ¿Qué pasa si se elimina una categoría con productos asociados? Se propone
   `ON DELETE RESTRICT` y baja lógica de categorías.
3. ¿Se pone un máximo de categorías por producto? Hoy no hay `maxItems`; se puede
   agregar después sin romper compatibilidad.
4. ¿La respuesta debe incluir los nombres de las categorías o solo los UUID?
   Hoy solo UUID, para no acoplar el contrato.
5. ¿El código del departamento viene de una tabla local o de otro servicio?

Ver también [`modos-de-falla.md`](./modos-de-falla.md).
