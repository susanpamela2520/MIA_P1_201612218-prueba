package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

/* Struct que contiene los atributos del disco. La primera letra de cada atributo se escribe en mayuscula para que sean
publicos ya que de este forma no muestra un error al querer acceder a la info. desde otro archivos. */
type MBR struct {
	Tamanio       int64
	FechaCreacion [20]byte
	NumeroRandom  int16
	TipoAjuste    byte
	Particiones   [4]Particion
}

var identificador int16 = 1

/*  Metodo que creara el disco. */
func crearDisco(tamanio int64, ajuste string, unidades string, ruta string) {

	// Llenando los datos del Disco.
	particionVacia := Particion{Inicio: -1}
	discoAux := MBR{
		Tamanio:      obtenerTamanioDisco(tamanio, unidades),
		NumeroRandom: identificador,
		TipoAjuste:   ajuste[0],
		Particiones: [4]Particion{
			particionVacia,
			particionVacia,
			particionVacia,
			particionVacia}}

	copy(discoAux.FechaCreacion[:], obtenerFecha())

	// Creando las carpetas necesarias para el alojamiento del disco.
	exec.Command("mkdir", "-p", ruta).Output()
	exec.Command("rmdir", ruta).Output()

	// Verificamos la existencia del archivo
	if _, err := os.Stat(ruta); err == nil {
		fmt.Println("El archivo ya existe, vuelva a intentarlo...")
		return
	}

	// Creando el archivo binario que simula el disco.
	archivo, _ := os.Create(ruta)

	defer archivo.Close()

	/*  
		1.Inicializar un buffer para la escritura, se escribe un 0 en el buffer y luego en el archivo.
		2.Crea un buffer con el contenido de  bytes.NewBuffer(contenido).
		3.Pasa los datos a binarios.
	*/ 
	buffer := bytes.NewBuffer([]byte{})               
	binary.Write(buffer, binary.BigEndian, uint8(0))  
	archivo.Write(buffer.Bytes())

	/* 
		1.Corrimiento del puntero del archivo para alcanzar el tamaño especificado.
		2.Escribiento la informacion.
	*/
	archivo.Seek(discoAux.Tamanio-int64(1), 0)
	archivo.Write(buffer.Bytes())

	// Posicionando el puntero en el inicio del archivo y limpiando el buffer de escritura.
	archivo.Seek(0, 0)
	buffer.Reset()

	// Escribiendo el struct que representa el mbr en el archivo.
	binary.Write(buffer, binary.BigEndian, &discoAux)
	archivo.Write(buffer.Bytes())

	identificador++

	fmt.Println("¡ EL disco se a creado exitosamente !\n")

}

/*  Metodo que elimina el disco, si el disco no existe muestra un error. */
func eliminarDisco(ruta string) {

	fmt.Println("¿Esta seguro que desea eliminar el disco (si/no)?")
	condicion := "no"
	fmt.Scanln(&condicion)

	if condicion == "si"{
		error := os.Remove(ruta)
		if error != nil {
			fmt.Println("¡ Error al eleminar el disco, intentelo nuevamente !\n")
		} else {
			fmt.Println("¡ Disco eliminado exitosamente !\n")
		}
	}else{
		fmt.Println("Operacion cancelada con exito...\n")
	}
}


func obtenerDisco(path string) *os.File {
	if _, err := os.Stat(path); err == nil {
		archivo, _ := os.OpenFile(path, os.O_RDWR, 0644)
		return archivo
	}
	return nil
}

/* Funcion que retorna la fecha y hora en que se creo el disco. */
func obtenerFecha() string {
	/* 
		1.Obtenemos la hora actual del sistema.
		2.Le damos formato a la cadena.
	*/
	tiempo := time.Now()

	fecha := fmt.Sprintf("%02d-%02d-%d %02d:%02d:%02d", tiempo.Day(), tiempo.Month(), tiempo.Year(),
		tiempo.Hour(), tiempo.Minute(), tiempo.Second())

	return fecha

}

/*  Funcion que retorna el tamaño del disco. */
func obtenerTamanioDisco(tamanio int64, unidades string) int64 {

	if (strings.Compare(unidades, "k")) == 0 {
		return int64(tamanio * 1024)
	} else if (strings.Compare(unidades, "m")) == 0 {
		return int64(tamanio * 1048576)
	}
	return int64(-1)
}
