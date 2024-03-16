package main

/* Importacion de librerias. */
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var contador = 0;
var mensaje = "";

/* Metodo de inicio donde se toma una cadena de entrada para posteriormente analizarla. */
func inicio() {
	lector := bufio.NewReader(os.Stdin)
	fmt.Println("\n******************* MIA PROYECTO 1 *******************")
	for {

		fmt.Print("Ingrese un comando : ")
		entrada, _ := lector.ReadString('\n')
		entrada = strings.ReplaceAll(entrada, "\n", "")
		entrada = strings.ReplaceAll(entrada, "\r", "")
		analizardor(strings.Split(entrada, " "))
	}
}

/* Metodo que analiza que comando es que se tiene que analizar */
func analizardor(comando []string) {

	if strings.Contains(comando[0], "#") {
		fmt.Println(comando)
	} else {

		cadena := strings.ToLower(comando[0])
		switch cadena {
		case "mkdisk":
			fmt.Println("CREACION DE DISCO")
			comandoMkdisk(comando)
		case "rmdisk":
			fmt.Println("ELIMINACION DE DISCO")
			comandoRmdisk(comando)
		case "fdisk":
			fmt.Println("CREACION PARTICIONES\n")
			comandoFdisk(comando)
		case "mount":
			fmt.Println("MONTAR PARTICIONES")
			comandoMount(comando)
		case "unmount":
			fmt.Println("DESMONTAR PARTICIONES")
			comandoUnmount(comando)
		case "mkfs":
			fmt.Println("SISTEMA DE ARCHIVOS")
			comandoMkfs(comando)
		case "login":
			fmt.Println("INICIAR SESION\n")
			comandoLogin(comando)
		case "logout":
			fmt.Println("CERRAR SESION\n")
			comandoLogout(comando)
		case "mkgrp":
			fmt.Println("CREAR GRUPOS\n")
			comandoMkgrp(comando)
		case "rmgrp":
			fmt.Println("ELIMINAR GRUPO\n")
			comandoRmgrp(comando)
		case "mkusr":
			fmt.Println("CREAR USUARIOS\n")
			comandoMkuser(comando)
		case "rmusr":
			fmt.Println("ELIMINAR USUARIO\n")
			comandoRmusr(comando)
		case "mkdir":
			fmt.Println("CREANDO CARPETAS\n")
			comandoMkdir(comando)
		case "rep":
			fmt.Println("REPORTES\n")
			comandoReporte(comando)
		case "pause":
			fmt.Println("PAUSE")
			pause()
		case "execute":
			fmt.Println("LEYENDO ARCHIVO")
			comandoExec(comando)
		case "exit":
			fmt.Println("ADIOS")
			exit()
		default:
			fmt.Println("Comando [" + cadena + "] no existe, vuelva a intentarlo..\n")

		}
	}
}

/* Metodo que analiza los parametros del comando MKDISK. */
func comandoMkdisk(comando []string) {

	tamanio := int64(-1)
	ajuste := ""
	unidad := ""
	letra := obtenerLetra(contador)
	ruta := obtenerRuta("MIA/P1/" + letra + ".dsk")
	contador++;

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-size":
			tamanio = obtenerTamanio(valor[1])
		case "-fit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "ff") == 0) || (strings.Compare(aux2, "bf") == 0) || (strings.Compare(aux2, "wf") == 0) {
				ajuste = aux2
			}

		case "-unit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "k") == 0) || (strings.Compare(aux2, "m") == 0) {
				unidad = aux2
			}

		default:
			error(aux1)
		}
	}

	if ajuste == "" {
		ajuste = "ff"
	}

	if unidad == "" {
		unidad = "m"
	}

	if tamanio > 0 && ruta != "" {
		crearDisco(tamanio, ajuste, unidad, ruta)
	} else {
		fmt.Println("¡ Faltan parametros obligatorios [MKDISK], vuelva a intentarlo !\n")
	}
}

/* Metodo que analiza los parametros del comando RMDISK. */
func comandoRmdisk(comando []string) {

	ruta := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-driveletter":
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
		default:
			error(aux1)
		}
	}

	if ruta != "" {
		eliminarDisco(ruta)
	} else {
		fmt.Println("¡ Faltan parametros obligatorios [RMDISK], vuelva a intentarlo !\n")
	}
}

/* Metodo que analiza los parametros del comando FDISK. */
func comandoFdisk(comando []string) {

	tamanio := int64(-1)
	unidad := ""
	ruta := ""
	tipoParticion := ""
	ajuste := ""
	nombre := ""
	borra := false
	agregar := false
	tipoEliminar := ""
	valorAgregar := int64(-1)

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-size":
			tamanio = obtenerTamanio(valor[1])
		case "-unit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "k") == 0) || (strings.Compare(aux2, "m") == 0) || (strings.Compare(aux2, "b") == 0) {
				unidad = aux2
			}

		case "-driveletter":
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
			fmt.Println("ruta ", ruta)
		case "-type":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "p") == 0) || (strings.Compare(aux2, "e") == 0) || (strings.Compare(aux2, "l") == 0) {
				tipoParticion = aux2
			}

		case "-fit":

			aux2 := strings.ToLower(valor[1])
			if (strings.Compare(aux2, "ff") == 0) || (strings.Compare(aux2, "bf") == 0) || (strings.Compare(aux2, "wf") == 0) {
				ajuste = aux2
			}

		case "-name":
			nombre = valor[1]
		case "-delete":
			tipoEliminar = valor[1]
			borra = true
		case "-add":
			valorAgregar = obtenerTamanio(valor[1])
			agregar = true
		default:
			error(aux1)
		}
	}

	if tamanio > 0 && ruta != "" && nombre != ""{
		if unidad == "" {
			unidad = "k"
		}
	
		if tipoParticion == "" {
			tipoParticion = "p"
		}
	
		if ajuste == "" {
			ajuste = "wf"
		}

		insertarParticion(tamanio, unidad, ruta, tipoParticion, ajuste, nombre)

	}else if nombre != "" && ruta != "" && borra {
		borrarParticion(tipoEliminar, ruta, nombre)
	} else if nombre != "" && ruta != "" && agregar{
		agregarParticion(ruta, nombre, valorAgregar, unidad)
	}else {
		fmt.Println("¡ Faltan parametros obligatorios [FDISK], vuelva a intentarlo !")
	}
}

/* Metodo que analiza los parametros del comando MOUNT. */
func comandoMount(comando []string) {

	ruta := ""
	nombre := ""
	letraDisco := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-driveletter":
			letraDisco = valor[1]
			ruta = obtenerRuta("MIA/P1/" + valor[1] + ".dsk")
		case "-name":
			nombre = valor[1]
		default:
			error(aux1)
		}
	}

	if ruta != "" && nombre != "" {
		montarParticion(ruta, nombre, letraDisco)
	} else {
		fmt.Println("¡ Faltan parametros obligatorios [MOUNT], vuelva a intentarlo !")
	}
}

/* Metodo que analiza los parametros del comando Unmount */
func comandoUnmount(comando []string){
	identificador := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-id":
			identificador = valor[1]
		default:
			error(aux1)
		}
	}

	if identificador != "" {
		desmotar(identificador)
	} else {
		fmt.Println("¡ Faltan parametros obligatorios [UNMOUNT], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando MKFS. */
func comandoMkfs(comando []string){


	identificador := ""
	tipoFormateo := ""
	tipoSistema := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-id":
			identificador = valor[1]
		case "-type":
			tipoFormateo = strings.ToLower(valor[1])
		case "-fs":
			tipoSistema = valor[1]
		default:
			error(aux1)
		}
	}

	if tipoFormateo == "" {
		tipoFormateo = "full"
	}

	if tipoSistema == "" {
		tipoSistema = "2fs"
	}

	if identificador != "" && tipoSistema == "2fs"{
		crearSistemaArchivosEXT2(identificador, tipoFormateo)
	}else if identificador != "" && tipoSistema == "3fs"{
		crearSistemaArchivosEXT3(identificador, tipoFormateo)
	} else{
		fmt.Println("¡ Faltan parametros obligatorios [MKFS], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando LOGIN. */
func comandoLogin(comando []string){

	usuario := ""
	contrasenia := ""
	id := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		case "-pass":
			contrasenia = strings.ToLower(valor[1])
		case "-id":
			id = valor[1]
		default:
			error(aux1)
		}
	}

	if id != "" && usuario != "" && contrasenia != "" {
		iniciarSesion(usuario, contrasenia, id)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [LOGIN], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando LOGOUT. */
func comandoLogout(comando []string){
	cerrarSesion()
}

/*  Metodo que analiza los parametros del comando MKGRP. */
func comandoMkgrp (comando[]string){
	
	grupo := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			grupo = valor[1]
		default:
			error(aux1)
		}
	}

	if grupo != "" {
		crearGrupo(grupo)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [MKGRP], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando RMGRP. */
func comandoRmgrp (comando[]string){
	
	grupo := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			grupo = valor[1]
		default:
			error(aux1)
		}
	}

	if grupo != "" {
		eliminarGrupo(grupo)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [RMGRP], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando MKUSER. */
func comandoMkuser(comando []string){

	usuario := ""
	contrasenia := ""
	grupoPertenece := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		case "-pass":
			contrasenia = strings.ToLower(valor[1])
		case "-grp":
			grupoPertenece = valor[1]
		default:
			error(aux1)
		}
	}

	if grupoPertenece != "" && usuario != "" && contrasenia != "" {
		crearUsuario(usuario, contrasenia, grupoPertenece)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [MKUSR], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando RMUSR. */
func comandoRmusr(comando[]string){

	usuario := ""
	
	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-user":
			usuario = valor[1]
		default:
			error(aux1)
		}
	}

	if usuario != "" {
		eliminarUsuario(usuario)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [RMUSR], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando MKDIR. */
func comandoMkdir(comando[]string){

	ruta := ""
	padre := false

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-path":
			ruta = valor[1]
		case "-r":
			padre = true
		default:
			error(aux1)
		}
	}

	if ruta != "" {
		crearCarpeta(ruta, padre)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [MKDIR], vuelva a intentarlo !")
	}
}

/*  Metodo que analiza los parametros del comando REP. */
func comandoReporte(comando []string){

	ruta := ""
	identificador := ""
	nombre := ""
	rutaOpcional := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")

		if strings.Contains(valor[0], "#") {
			break
		}

		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-name":
			nombre = valor[1]
		case "-path":
			ruta = obtenerRuta(valor[1])
		case "-id":
			identificador = valor[1]
		case "-ruta":
			rutaOpcional = valor[1]
		default:
			error(aux1)
		}
	}

	if nombre != "" && ruta != "" && identificador != "" {
		analizarReporte(nombre, ruta, identificador, rutaOpcional)
	}else{
		fmt.Println("¡ Faltan parametros obligatorios [REP], vuelva a intentarlo !")
	}
}

/* Metodo que analiza el comando exec para ejecutar el archivo. */
func comandoExec(comando []string) {

	ruta := ""

	for iterador := 1; iterador < len(comando); iterador++ {

		valor := strings.Split(comando[iterador], "=")
		aux1 := strings.ToLower(valor[0])

		switch aux1 {
		case "-path":
			ruta = obtenerRuta(valor[1])
		default:
			error(aux1)
		}
	}

	if ruta != "" {
		leerArchivoEntrada(ruta)
	}
}

/* Metodo que abre y analiza el archivo para la ejecucion de los comandos. */
func leerArchivoEntrada(ruta string) {

	archivo, error := os.Open(ruta)

	if error != nil {
		fmt.Println("¡ Error, el archivo no existe, vuelva a intentarlo !")
		inicio()
		archivo.Close()
	}

	scanner := bufio.NewScanner(archivo)

	for scanner.Scan() {

		linea := scanner.Text()

		concatenar := linea

		if concatenar != "" {

			if strings.Contains(string(concatenar[0]), "#") {
				fmt.Println(concatenar)
				concatenar = ""
			} else {
				fmt.Println(concatenar)
				analizardor(strings.Split(concatenar, " "))
				concatenar = ""
			}

		}
	}
}

/* Funcion para obtener limpia la ruta. */
func obtenerRuta(valor string) string {

	ruta := ""
	if strings.Contains(valor, "\"") {
		ruta = strings.ReplaceAll(valor, "\"", "")
	} else {
		ruta = valor
	}

	return ruta
}

/* Funcion que retorna el tamaño del disco. */
func obtenerTamanio(valor string) int64 {

	tamanio, _ := strconv.Atoi(valor)
	if tamanio > 0 {
		return int64(tamanio)
	}
	return -1
}

/* Metodo para reportar algun comando invalido */
func error(comando string) {
	if comando != "" {
		fmt.Println("El comando ["+comando+"] no es reconocido, ingrese un comando valido.")
	}
}

/* Metodo que pausa la ejecucion del programa. */
func pause() {
	fmt.Println("Presione cualquier tecla para continuar .....")
	tecla := ""
	fmt.Scanln(&tecla)
}

/* Metodo que finaliza la ejecucion del programa. */
func exit() {
	fmt.Println("¡ Finalizacion del programa realizado exitosamente !")
	os.Exit(0)
}