package main

/*
	github.com/gorilla/mux: Librería que implementa un router de solicitudes y un despachador
	que permite que coincidan las solicitudes entrantes con su respectivo controlador.
*/

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

var particion string;

/* 
	Struct donde recibiremos los datos enviados desde el frontend, las etiquetas `json:"cualquierNombre"` proporcionan informacion adicional, en este caso 
	para parsear el JSON
*/
type usuario struct {
	Usuario string `json:"Usuario"`
	Contrasenia string `json:"Contrasenia"`
	IdParticion string `json:"IdParticion"`
}

func misRutas() *mux.Router{
	enrutador := mux.NewRouter().StrictSlash(true)
	enrutador.HandleFunc("/analizar", analizarArchivo).Methods("POST")
	enrutador.HandleFunc("/iniciarSesion", validarDatos).Methods("POST")
	enrutador.HandleFunc("/cerrarSesion", finalizarCesion).Methods("POST")
	enrutador.HandleFunc("/obtenerImg", obtenerImg).Methods("GET")

	return enrutador
}

/* 
	Funcion que recibe el nombre del archivo desde el frontend y lo envia para empezar a realizar el 
    sistema de archivos. 
*/
func analizarArchivo(respuesta http.ResponseWriter, peticion *http.Request) {
	
	reqBody, err := ioutil.ReadAll(peticion.Body)
	if err != nil {
		fmt.Fprintf(respuesta, "Ocurrio un error, vuelva a intentarlo..")
	}

	
	nombreArchivo := string(reqBody)
	ruta := "-path=/home/" + nombreArchivo
	comando := [2]string{"exec",ruta}
	
	analizardor(comando[:])
	
	respuesta.Header().Set("Content-type", "application/json")
	json.NewEncoder(respuesta).Encode(mensaje)		
}

/* 
	Funcion que analiza los datos ingresados por el usuario para iniciar sesion.
*/
func validarDatos(respuesta http.ResponseWriter, peticion *http.Request){
	reqBody, err := ioutil.ReadAll(peticion.Body)
	if err != nil {
		fmt.Fprintf(respuesta, "Ocurrio un error, vuelva a intentarlo..")
	}
	
	var iniciando usuario;
	json.Unmarshal(reqBody, &iniciando)
	respuesta.Header().Set("Content-type", "application/json")
	particion = iniciando.IdParticion
	encontrado, validacion := iniciarSesion(iniciando.Usuario, iniciando.Contrasenia, iniciando.IdParticion)
	if !encontrado {
		respuesta.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(respuesta).Encode(validacion)	
		return
	}
	
	json.NewEncoder(respuesta).Encode(validacion)	
}

/* 
	Funcion que analiza los datos ingresados por el usuario para iniciar sesion.
*/
func finalizarCesion(respuesta http.ResponseWriter, peticion *http.Request){
	
	existeSesion, validacion := cerrarSesion()
	if !existeSesion {
		respuesta.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(respuesta).Encode(validacion)	
		return
	}
	respuesta.Header().Set("Content-type", "application/json")	
	json.NewEncoder(respuesta).Encode(validacion)	
}


func obtenerImg(respuesta http.ResponseWriter, peticion *http.Request){

	var aux [25]string
	for i := 0; i < 25; i++ {
		fmt.Println(arregloReportes[i].Ruta)
	}
	for i := 0; i < 25; i++ {
		if particion == arregloReportes[i].IdParticion {
			aux[i] = arregloReportes[i].Ruta
		}
	}
	respuesta.Header().Set("Content-type", "application/json")	
	json.NewEncoder(respuesta).Encode(aux)	
}