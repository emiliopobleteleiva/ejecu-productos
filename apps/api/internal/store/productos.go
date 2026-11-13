// Package store contiene el acceso a PostgreSQL.
package store

// AlmacenProductos define la persistencia.
//
// TODO Hito 2: consultas sqlc y transacciones.
// TODO Hito 2: insertar el producto y sus filas en producto_categorias dentro
// de una misma transacción.
// TODO Hito 2: manejar la violación de UNIQUE(sku) y traducirla a un error de
// dominio que el service convierta en 409.
// TODO Hito 2: crear la tabla puente
// producto_categorias(producto_id, categoria_id) con PK compuesta y FKs.
// TODO Hito 2: guardar costo como bigint, fechas como timestamptz e IDs como
// uuid.
type AlmacenProductos interface{}
