// Package service contiene las reglas de negocio del dominio de productos.
package service

// ServicioProductos define las reglas de negocio.
//
// TODO Hito 2: Crear, Listar, Obtener, Actualizar y Desactivar.
// TODO Hito 2: validar SKU único coordinando con store (consulta previa y
// traducción de la violación UNIQUE de PostgreSQL a un error de dominio).
// TODO Hito 2: verificar que cada categoria_id exista, ya sea contra una
// tabla local de categorías o contra la API del servicio dueño.
// TODO Hito 2: validar que el costo sea entero y mayor o igual a cero.
// TODO Hito 2: aplicar la regla de permite_venta en el flujo de venta.
// TODO Hito 2: implementar la baja lógica y decidir si un producto inactivo
// puede reactivarse.
type ServicioProductos interface{}
