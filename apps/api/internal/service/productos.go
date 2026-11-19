// Package service contiene las reglas de negocio del dominio de productos.
package service

// ServicioProductos define las reglas de negocio.
//
// TODO Hito 2: Crear, Listar, Obtener, Actualizar y Desactivar.
// TODO Hito 2: generar el SKU con el código del departamento y un secuencial
// por departamento, dentro de la misma transacción del insert, y traducir la
// violación de UNIQUE(sku) a un error de dominio.
// TODO Hito 2: verificar que el departamento exista antes de generar el SKU.
// TODO Hito 2: verificar que cada categoria_id exista, ya sea contra una
// tabla local de categorías o contra la API del servicio dueño.
// TODO Hito 2: validar costo_unitario entero y mayor o igual a cero, y que
// tipo_impuesto sea IVA o EXENTO.
// TODO Hito 2: impedir cambios de sku y departamento_id (son inmutables).
// TODO Hito 2: aplicar la regla de permite_venta en el flujo de venta.
// TODO Hito 2: implementar la baja lógica y decidir si un producto inactivo
// puede reactivarse.
type ServicioProductos interface{}
