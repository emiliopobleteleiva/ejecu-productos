# Decisiones de modelado

## Categorías múltiples por producto (Opción A)

**Estado:** adoptada para el contrato del Hito 1. La persistencia queda para el
Hito 2.

### Contexto

Un producto puede pertenecer a más de una categoría. El contrato debe permitir
esa relación sin obligar a los consumidores a cambiar cuando se agregue una
segunda categoría.

### Opciones evaluadas

| Opción | Modelo | Ventajas | Desventajas |
|---|---|---|---|
| A (adoptada) | `categoria_ids []string` en el producto y tabla puente `producto_categorias(producto_id, categoria_id)` | Relación N:N normalizada, filtros por categoría, sin duplicar productos | Requiere tabla puente y validar la existencia de cada UUID |
| B | `categoria_id` singular en el producto | Más simple de implementar | Limita a una categoría; migrar después rompe el contrato |
| C | Objetos de categoría anidados en la respuesta | Cómodo para el cliente | Acopla el contrato a los datos de otro servicio y puede quedar obsoleto |

### Decisión

- En el contrato: `categoria_ids []string`, obligatorio, `minItems: 1` y
  `uniqueItems: true`.
- En la base de datos del Hito 2: tabla puente
  `producto_categorias(producto_id, categoria_id)` con PK compuesta y FKs.
- No existe `categoria_id` singular dentro de `Producto`. El único
  `categoria_id` singular es el query param de filtro del listado.
- La existencia de cada categoría se valida en el service (tabla local o API
  del servicio dueño de categorías). El contrato solo valida el formato UUID.

### Preguntas abiertas para el equipo

1. ¿La categoría es un dato propio de este servicio o vive en otro servicio?
   Si vive en otro, ¿se valida por API en cada escritura o se mantiene una
   copia local?
2. ¿Qué pasa si se elimina una categoría con productos asociados? Se propone
   `ON DELETE RESTRICT` y baja lógica de categorías, para no dejar productos
   sin categoría.
3. ¿Se pone un máximo de categorías por producto? Hoy el contrato no tiene
   `maxItems`; se puede agregar más adelante sin romper compatibilidad.
4. ¿La respuesta debe incluir los nombres de las categorías o solo los UUID?
   Por ahora solo UUID, para no acoplar el contrato.
5. ¿El listado debe incluir productos de categorías inactivas? Por ahora no se
   filtra por estado de categoría porque ese dato no pertenece a este servicio.

Ver también [`modos-de-falla.md`](./modos-de-falla.md).
