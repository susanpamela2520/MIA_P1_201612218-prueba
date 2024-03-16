package main

/*
	Log: Librería que maneja el registro de las operaciones o bitácora de cambios.
	Cors: mecanismo que permite que se puedan solicitar recursos restringidos en una página web desde un
	dominio diferente del dominio que sirvió el primer recurso.
	net/http: Librería que permite realizar la funcionalidad del servidor, como por ejemplo la realización
	de peticiones y devolución de respuestas del servidor.
*/
import (
	"fmt"
	"log"
	"net/http"
	"github.com/rs/cors"
)

/* Metodo que inicia el servidor. */
func iniciarServidor(){
	/*  */
	cors := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
	})
	fmt.Println("El server corriendo a tope bro!!")
	/* Variable que guarda obtiene cada ruta que el cliente solicita. */
	rutas := misRutas()

	/* Con "log" ponemos a escuchar el servidor */
	log.Fatal(http.ListenAndServe(":4000", cors.Handler(rutas)))

}