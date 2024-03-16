package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

type Reporte struct {
	IdParticion    string
	Ruta string
}

var arregloReportes[25] Reporte;

/* Metodo que analiza la cadena que ingresa y va al metodo solicitado para generar el reporte */
func analizarReporte(nombre string, rutaReporte string, id string, rutaOpcional string) {

	nombreReporte := strings.ToLower(nombre)

	// Creando las carpetas necesarias para el alojamiento del disco.
	exec.Command("mkdir", "-p", rutaReporte).Output()
	exec.Command("rmdir", rutaReporte).Output()

	switch nombreReporte {
	case "mbr":
		fmt.Println("REPORTE MBR")
		reporteMbr(rutaReporte, id)
	case "disk":
		fmt.Println("REPORTE DEL DISCO")
		reporteDisk(rutaReporte, id)
	case "tree":
		fmt.Println("REPORTE ARBOL DEL SISTEMA")
		reporteTree(rutaReporte, id)
	case "file":
		fmt.Println("REPORTE ARCHIVO\n")
		reporteFile(rutaReporte, id, rutaOpcional)
	case "inode":
		fmt.Println("REPORTE INODO")
		reporteInodo(rutaReporte, id)
	case "block":
		fmt.Println("REPORTE BLOQUE")
		reporteBlock(rutaReporte, id)
	case "bm_inode":
		fmt.Println("REPORTE BITMAP BLOQUE")
		reporteMapBloque(rutaReporte, id)
	case "bm_bloc":
		fmt.Println("REPORTE BITMAP INODO")
		reporteMapInodo(rutaReporte, id)
	case "sb":
		fmt.Println("REPORTE SUPER BLOQUE")
		reporteSuperBloque(rutaReporte, id)
	default:
		error(nombreReporte)
	}
}

/* Metodo que genera el reporte mbr. */
func reporteMbr(rutaReporte string, id string) {

	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)


	grafoDisco := "<TR>\n"
	grafoDisco += "<TD> Tamaño MBR </TD>\n"
	grafoDisco +=  "<TD>"+ strconv.Itoa(int(discoAux.Tamanio)) +"</TD>\n"
	grafoDisco +=  "</TR>\n"
	grafoDisco +=   "<TR>\n"
	grafoDisco +=  "<TD> Fecha Creacion </TD>\n"
	grafoDisco +=  "<TD>"+ cadenaLimpia(discoAux.FechaCreacion[:]) +"</TD>\n"
	grafoDisco +=  "</TR>\n"
	grafoDisco +=  "<TR>\n"
	grafoDisco +=  "<TD> Identificador </TD>\n"
	grafoDisco +=  "<TD>"+ strconv.Itoa(int(discoAux.NumeroRandom)) +"</TD>\n"
	grafoDisco +=  "</TR>\n"
	grafoDisco +=  "<TR>\n"
	grafoDisco += "<TD> Tipo ajuste </TD>\n"
	grafoDisco +=  "<TD>"+ string(discoAux.TipoAjuste) +"</TD>\n"
	grafoDisco += "</TR>\n";

	grafoParticiones := ""

	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1{
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD><B> Nombre Particion </B></TD>\n"
			grafoParticiones += "<TD>"+ cadenaLimpia(discoAux.Particiones[i].Nombre[:]) +"</TD>\n"
			grafoParticiones += "</TR>\n"
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD> Tipo Particion </TD>\n"
			grafoParticiones += "<TD>"+ string(discoAux.Particiones[i].Tipo) +"</TD>\n"
			grafoParticiones += "</TR>\n"
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD> Estatus Particion </TD>\n"
			grafoParticiones += "<TD>"+ string(discoAux.Particiones[i].Estado) +"</TD>\n"
			grafoParticiones += "</TR>\n"
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD> Ajuste Particion </TD>\n"
			grafoParticiones += 	"<TD>"+ string(discoAux.Particiones[i].Ajuste) +"</TD>\n"
			grafoParticiones += "</TR>\n"
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD> Inicio Particion </TD>\n"
			grafoParticiones += "<TD>"+ strconv.Itoa(int(discoAux.Particiones[i].Inicio)) +"</TD>\n"
			grafoParticiones += "</TR>\n"
			grafoParticiones += "<TR>\n"
			grafoParticiones += "<TD> Tamaño Particion </TD>\n"
			grafoParticiones += "<TD>"+ strconv.Itoa(int(discoAux.Particiones[i].Tamanio)) +"</TD>\n"
			grafoParticiones += "</TR>\n";	
		}
	}

	dot := "digraph MBR{ \n"
    dot += "node [shape=plaintext]; \n"
	dot += "tabla [label=< \n"
	dot += "<TABLE ALIGN=\"LEFT\"> \n"
	dot += "<TR> \n"
	dot += "<TD><B> Descripcion </B></TD> \n"
	dot += "<TD><B> Valor  </B></TD> \n"
	dot += "</TR> \n"
	dot += grafoDisco
	dot += grafoParticiones
	dot += "</TABLE> \n"
	dot +=  ">]; \n"
	dot += "}";

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(dot)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println(" ¡ Reporte generado exitosamente !")
}

/* Metodo que genera el reporte disk. */
func reporteDisk(rutaReporte string, id string) {

	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	tamanioDisponible := float64(discoAux.Tamanio)

	grafo := "digraph Disk {\n"
	grafo = grafo + "graph [ratio = fill]; \n "
	grafo = grafo + "node  [label=\"N\", fontsize=15, shape=plaintext]; \n "
	grafo = grafo + "graph [bb=\"0,0,352,154\"]; \n"
	grafo = grafo + "arset [label=< \n"
	grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
	grafo = grafo + "<TR> \n"
	grafo = grafo + "<TD> MBR </TD> \n"
	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Inicio != -1 {

			tipo := string(discoAux.Particiones[i].Tipo)
			tamanioDisponible -= float64(discoAux.Particiones[i].Tamanio)
			porcentaje := float64((float64(discoAux.Particiones[i].Tamanio) * 100) / float64(discoAux.Tamanio))
			porcentajeConvertido := fmt.Sprintf("%.3f", porcentaje)
			if discoAux.Particiones[i].Tipo == 'p' || discoAux.Particiones[i].Tipo == 'e' {
				grafo = grafo + "<TD> <TABLE BORDER=\"0\"> \n"
				grafo = grafo + "<TR><TD>" + cadenaLimpia(discoAux.Particiones[i].Nombre[:]) + "</TD></TR> \n"
				grafo = grafo + "<TR><TD>" + tipo + "</TD></TR> \n"
				grafo = grafo + "<TR><TD>" + porcentajeConvertido + "%</TD></TR> \n"
				grafo = grafo + "</TABLE> </TD>; \n"
			} else {
				fmt.Println("LOGICA")
			}
		}
	}

	if tamanioDisponible > 0 {
		porcentaje := float64((float64(tamanioDisponible) * 100) / float64(discoAux.Tamanio))
		porcentajeConvertido := fmt.Sprintf("%.3f", porcentaje)
		grafo = grafo + "<TD> <TABLE BORDER=\"0\"> \n"
		grafo = grafo + "<TR><TD> LIBRE </TD></TR> \n"
		grafo = grafo + "<TR><TD>" + porcentajeConvertido + "%</TD></TR> \n"
		grafo = grafo + "</TABLE> </TD>; \n"
	}

	grafo = grafo + "</TR> \n"
	grafo = grafo + "</TABLE> \n"
	grafo = grafo + ">]; \n"
	grafo = grafo + "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(grafo)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println("¡ Reporte generado exitosamente !")
}

func reporteInodo(rutaReporte string, id string){
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)

	inicioBitMapInodos := int64(super.InicioBitMapsInodos)
	aux := uint8(1)
	contador := int64(0)

	grafo := "digraph Inodo {\n"
	grafo += "node [shape=plaintext]; \n"

	for {
		if aux == uint8(1) {
			lectura := make([]byte, int(unsafe.Sizeof(aux)))
			archivo.Seek(inicioBitMapInodos+contador, 0)
			archivo.Read(lectura)
			buffer := bytes.NewBuffer(lectura)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}


	for i := 0; i < int(contador); i++ {

		iNodo := TablaInodos{}
		contenido := make([]byte, int(unsafe.Sizeof(iNodo)))
		posicionInodo := int64(super.InicioTablaInodos) + (int64(i) * int64(unsafe.Sizeof(TablaInodos{})))
		archivo.Seek(posicionInodo, 0)
		archivo.Read(contenido)
		buffer := bytes.NewBuffer(contenido)
		binary.Read(buffer, binary.BigEndian, &iNodo)

		grafo = grafo + "Inodo" + strconv.Itoa(int(posicionInodo)) + "[shape=plaintext];\n "
		grafo = grafo + "Inodo" + strconv.Itoa(int(posicionInodo)) + "[label=< \n"
		grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Inodo </B></TD> \n"
		grafo = grafo + "<TD> " + strconv.Itoa(int(i)) + " </TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> UID </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.IdUsuario)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> GUID </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.IdGrupo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Tamaño </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.TamanioArchivo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Lectura </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaLectura[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Creacion </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaCreacion[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Modificacion </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaModificacion[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		for i := 0; i < 15; i++ {
			if iNodo.Bloque[i] != -1 {
				grafo = grafo + "<TR> \n"
				grafo = grafo + "<TD><B> Apuntador " + strconv.Itoa(int(i)) + "</B></TD> \n"
				grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Bloque[i])) + "</TD> \n"
				grafo = grafo + "</TR> \n"
			} else {
				grafo = grafo + "<TR> \n"
				grafo = grafo + "<TD><B> Apuntador " + strconv.Itoa(int(i)) + "</B></TD> \n"
				grafo = grafo + "<TD> -1 </TD> \n"
				grafo = grafo + "</TR> \n"
			}
		}
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Tipo </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Tipo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Permisos </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Permisos)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "</TABLE> \n"
		grafo = grafo + ">]; \n"
	}
	grafo = grafo + "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(grafo)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println("¡ Reporte generado exitosamente !")
}

func reporteBlock(rutaReporte string, id string){
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)

	inicioBitMapInodos := int64(super.InicioBitMapsInodos)

	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {
			lectura := make([]byte, int(unsafe.Sizeof(aux)))
			archivo.Seek(inicioBitMapInodos+contador, 0)
			archivo.Read(lectura)
			buffer := bytes.NewBuffer(lectura)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}

	grafo := "digraph BLoque {\n"
	grafo += "rankdir = LR; \n"
	grafo += "graph[overlap = \"false\", splines = \"true\"]"

	for i := 0; i < int(contador); i++ {

		iNodo := TablaInodos{}
		contenido := make([]byte, int(unsafe.Sizeof(iNodo)))
		posicionInodo := int64(super.InicioTablaInodos) + (int64(i) * int64(unsafe.Sizeof(TablaInodos{})))
		archivo.Seek(posicionInodo, 0)
		archivo.Read(contenido)
		buffer := bytes.NewBuffer(contenido)
		binary.Read(buffer, binary.BigEndian, &iNodo)

		for i := 0; i < 15; i++ {
			if iNodo.Bloque[i] != -1 { // Si es un bloque de carpeta.
				if iNodo.Tipo == int64(0) {
					bloqueCarpeta := BloqueCarpeta{}
					contenidoCarpeta := make([]byte, int(unsafe.Sizeof(bloqueCarpeta)))
					archivo.Seek(iNodo.Bloque[i], 0)
					archivo.Read(contenidoCarpeta)
					buffer2 := bytes.NewBuffer(contenidoCarpeta)
					binary.Read(buffer2, binary.BigEndian, &bloqueCarpeta)

					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[shape=plaintext];\n "
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[label=< \n"
					grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> Bloque Carpeta </B></TD> \n"
					grafo = grafo + "<TD> " + strconv.Itoa(int(iNodo.Bloque[i])) + " </TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> B_NAME </B></TD> \n"
					grafo = grafo + "<TD><B> B_INODO </B></TD> \n"
					grafo = grafo + "</TR> \n"
					for j := 0; j < 4; j++ {
						grafo = grafo + "<TR> \n"
						grafo = grafo + "<TD>" + cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) + "</TD> \n"
						grafo = grafo + "<TD>" + strconv.Itoa(int(bloqueCarpeta.Contenidos[j].Apuntador)) + "</TD> \n"
						grafo = grafo + "</TR> \n"
					}
					grafo = grafo + "</TABLE> \n"
					grafo = grafo + ">]; \n"
					

				} else if iNodo.Tipo == int64(1) { // Si es un bloque de archivo.

					archivos := obtenerBloqueArchivo(archivo, iNodo.Bloque[i])
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[shape=plaintext];\n "
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[label=< \n"
					grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> Bloque Archivo </B></TD> \n"
					grafo = grafo + "<TD> " + strconv.Itoa(int(iNodo.Bloque[i])) + " </TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD>" + cadenaLimpia(archivos.Datos[:]) + "</TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "</TABLE> \n"
					grafo = grafo + ">]; \n"
				}
			}
		}
	}
	grafo = grafo + "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(grafo)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println("¡ Reporte generado exitosamente !")
}

func reporteMapBloque(rutaReporte string, id string){
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posisionBitMap := int64(super.InicioBitMapsBloques)
	aux := uint8(1)

	dot := "digraph BitBloques {\n"
	dot += "mapa[shape=plaintext]; \n"
	dot += "mapa[label=< \n"
	dot += "<TABLE border=\"3\" cellspacing=\"10\" cellpadding=\"10\"  bgcolor=\"white\" >\n"

	for i := 0; i < 200; i++{
		lectura := make([]byte, int(unsafe.Sizeof(aux)))
		archivo.Seek(posisionBitMap+int64(i), 0)
		archivo.Read(lectura)
		buffer := bytes.NewBuffer(lectura)
		binary.Read(buffer, binary.BigEndian, &aux)
		if i == 0{
			dot += "<TR>\n"
			if aux == uint8(1){
				dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD>\n"
			}else{
				dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD>\n"
			}
		}else{
			if i % 20 == 0{
				if aux == uint8(1){
					dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD> </TR>"
					dot +="<TR>\n"
				}else{
					dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD> </TR>"
					dot += "<TR>\n"
				}
			}else{
				if aux == uint8(1){
					dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD>\n"
				}else{
					dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD>\n"
				}
			}
		}
	}

	dot += "</TR> \n"
	dot += "</TABLE> \n"
	dot += ">] \n;"
	dot += "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(dot)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println("¡ Reporte generado exitosamente !")
}

func reporteMapInodo(rutaReporte string, id string){
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posisionBitMap := int64(super.InicioBitMapsInodos)
	aux := uint8(1)

	dot := "digraph BitInodo {\n"
	dot += "mapa[shape=plaintext]; \n"
	dot += "mapa[label=< \n"
	dot += "<TABLE border=\"3\" cellspacing=\"10\" cellpadding=\"10\"  bgcolor=\"white\" >\n"

	for i := 0; i < 200; i++{
		lectura := make([]byte, int(unsafe.Sizeof(aux)))
		archivo.Seek(posisionBitMap+int64(i), 0)
		archivo.Read(lectura)
		buffer := bytes.NewBuffer(lectura)
		binary.Read(buffer, binary.BigEndian, &aux)
		if i == 0{
			dot += "<TR>\n"
			if aux == uint8(1){
				dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD>\n"
			}else{
				dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD>\n"
			}
		}else{
			if i % 20 == 0{
				if aux == uint8(1){
					dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD> </TR>"
					dot +="<TR>\n"
				}else{
					dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD> </TR>"
					dot += "<TR>\n"
				}
			}else{
				if aux == uint8(1){
					dot += "<TD border=\"3\" bgcolor=\"grey\"> 1 </TD>\n"
				}else{
					dot += "<TD border=\"3\" bgcolor=\"yellow\"> 0 </TD>\n"
				}
			}
		}
	}

	dot += "</TR> \n"
	dot += "</TABLE> \n"
	dot += ">] \n;"
	dot += "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(dot)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println("¡ Reporte generado exitosamente !")
}

/* Metodo que genera el reporte tree. */
func reporteTree(rutaReporte string, id string) {

	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)

	inicioBitMapInodos := int64(super.InicioBitMapsInodos)

	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {
			lectura := make([]byte, int(unsafe.Sizeof(aux)))
			archivo.Seek(inicioBitMapInodos+contador, 0)
			archivo.Read(lectura)
			buffer := bytes.NewBuffer(lectura)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}

	grafo := "digraph Tree {\n"
	grafo += "rankdir = LR; \n"
	grafo += "graph[overlap = \"false\", splines = \"true\"]"

	for i := 0; i < int(contador); i++ {

		iNodo := TablaInodos{}
		contenido := make([]byte, int(unsafe.Sizeof(iNodo)))
		posicionInodo := int64(super.InicioTablaInodos) + (int64(i) * int64(unsafe.Sizeof(TablaInodos{})))
		archivo.Seek(posicionInodo, 0)
		archivo.Read(contenido)
		buffer := bytes.NewBuffer(contenido)
		binary.Read(buffer, binary.BigEndian, &iNodo)

		grafo = grafo + "Inodo" + strconv.Itoa(int(posicionInodo)) + "[shape=plaintext];\n "
		grafo = grafo + "Inodo" + strconv.Itoa(int(posicionInodo)) + "[label=< \n"
		grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Inodo </B></TD> \n"
		grafo = grafo + "<TD> " + strconv.Itoa(int(i)) + " </TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> UID </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.IdUsuario)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> GUID </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.IdGrupo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Tamaño </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.TamanioArchivo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Lectura </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaLectura[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Creacion </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaCreacion[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Modificacion </B></TD> \n"
		grafo = grafo + "<TD>" + cadenaLimpia(iNodo.FechaModificacion[:]) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		for i := 0; i < 15; i++ {
			if iNodo.Bloque[i] != -1 {
				grafo = grafo + "<TR> \n"
				grafo = grafo + "<TD><B> Apuntador " + strconv.Itoa(int(i)) + "</B></TD> \n"
				grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Bloque[i])) + "</TD> \n"
				grafo = grafo + "</TR> \n"
			} else {
				grafo = grafo + "<TR> \n"
				grafo = grafo + "<TD><B> Apuntador " + strconv.Itoa(int(i)) + "</B></TD> \n"
				grafo = grafo + "<TD> -1 </TD> \n"
				grafo = grafo + "</TR> \n"
			}
		}
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Tipo </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Tipo)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "<TR> \n"
		grafo = grafo + "<TD><B> Permisos </B></TD> \n"
		grafo = grafo + "<TD>" + strconv.Itoa(int(iNodo.Permisos)) + "</TD> \n"
		grafo = grafo + "</TR> \n"
		grafo = grafo + "</TABLE> \n"
		grafo = grafo + ">]; \n"

		for i := 0; i < 15; i++ {
			if iNodo.Bloque[i] != -1 { // Si es un bloque de carpeta.

				grafo = grafo + "Inodo" + strconv.Itoa(int(posicionInodo)) + ":" + strconv.Itoa(int(posicionInodo)) + "->" + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "\n"

				if iNodo.Tipo == int64(0) {

					bloqueCarpeta := BloqueCarpeta{}
					contenidoCarpeta := make([]byte, int(unsafe.Sizeof(bloqueCarpeta)))
					archivo.Seek(iNodo.Bloque[i], 0)
					archivo.Read(contenidoCarpeta)
					buffer2 := bytes.NewBuffer(contenidoCarpeta)
					binary.Read(buffer2, binary.BigEndian, &bloqueCarpeta)

					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[shape=plaintext];\n "
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[label=< \n"
					grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> Bloque Carpeta </B></TD> \n"
					grafo = grafo + "<TD> " + strconv.Itoa(int(iNodo.Bloque[i])) + " </TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> B_NAME </B></TD> \n"
					grafo = grafo + "<TD><B> B_INODO </B></TD> \n"
					grafo = grafo + "</TR> \n"
					for j := 0; j < 4; j++ {
						grafo = grafo + "<TR> \n"
						grafo = grafo + "<TD>" + cadenaLimpia(bloqueCarpeta.Contenidos[j].Nombre[:]) + "</TD> \n"
						grafo = grafo + "<TD>" + strconv.Itoa(int(bloqueCarpeta.Contenidos[j].Apuntador)) + "</TD> \n"
						grafo = grafo + "</TR> \n"
					}
					grafo = grafo + "</TABLE> \n"
					grafo = grafo + ">]; \n"
					for k := 0; k < 4; k++ {
						if (cadenaLimpia(bloqueCarpeta.Contenidos[k].Nombre[:]) != "." && cadenaLimpia(bloqueCarpeta.Contenidos[k].Nombre[:]) != "..") && bloqueCarpeta.Contenidos[k].Apuntador != int64(-1) {
							grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + ":" + strconv.Itoa(int(iNodo.Bloque[i])) + "->" + "Inodo" + strconv.Itoa(int(bloqueCarpeta.Contenidos[k].Apuntador)) + "\n"
						}
					}

				} else if iNodo.Tipo == int64(1) { // Si es un bloque de archivo.

					archivos := obtenerBloqueArchivo(archivo, iNodo.Bloque[i])
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[shape=plaintext];\n "
					grafo = grafo + "bloque" + strconv.Itoa(int(iNodo.Bloque[i])) + "[label=< \n"
					grafo = grafo + "<TABLE ALIGN=\"LEFT\"> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD><B> Bloque Archivo </B></TD> \n"
					grafo = grafo + "<TD> " + strconv.Itoa(int(iNodo.Bloque[i])) + " </TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "<TR> \n"
					grafo = grafo + "<TD>" + cadenaLimpia(archivos.Datos[:]) + "</TD> \n"
					grafo = grafo + "</TR> \n"
					grafo = grafo + "</TABLE> \n"
					grafo = grafo + ">]; \n"
				}
			}
		}
	}
	grafo = grafo + "}"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(grafo)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println(" ¡ Reporte generado exitosamente !")
}

/* Metodo qeu genera el reporte del super bloque. */
func reporteSuperBloque(rutaReporte string, id string){

	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)

	dot := "digraph SuperBloque{ \n"
	dot += "node [shape=plaintext]; \n"
	dot += "tabla [label=< \n"
	dot += "<TABLE ALIGN=\"LEFT\"> \n"
	dot += "<TR> \n"
	dot += 	   "<TD><B> Descripcion </B></TD> \n"
	dot += 	   "<TD><B> Valor  </B></TD> \n"
	dot += "</TR> \n"
	dot += "<TR>\n"
	dot += "<TD><B> Nombre Particion </B></TD>\n"
	dot += "<TD>"+ cadenaLimpia(nombreParticion[:]) +"</TD>\n"
    dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Sistema Archivos </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.IdSistema)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Numero Inodos </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroInodos)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Numero Bloques </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroBloques)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Numero Inodos Libres </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroInodosLibres)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Numero Bloques Libres </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroBloquesLibres)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Ultima Fecha Montado </B></TD>\n"
	dot += "<TD>"+ cadenaLimpia(super.UltimaFechaMontado[:]) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Ultima Fecha Desmontado </B></TD>\n"
	dot += "<TD>"+ cadenaLimpia(super.UltimaFechaMontado[:]) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> # Veces Montado Sistema </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroSistemaMontado)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Id Sistema </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.NumeroMagico)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Tamaño INodo </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.TamanioInodo)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Tamaño Bloque </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.TamanioBloque)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Primer Inodo Libre </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.PrimerInodoLibre)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Pimer Bloque Libre </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.PrimerBloqueLibre)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Inicio BitMap Inodo </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.InicioBitMapsInodos)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Inicio BitMap Bloque </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.InicioBitMapsBloques)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Inicio Tabla Inodo </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.InicioTablaInodos)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "<TR>\n"
	dot += "<TD><B> Inicio Tabla Bloque </B></TD>\n"
	dot += "<TD>"+ strconv.Itoa(int(super.InicioTablaBloques)) +"</TD>\n"
	dot += "</TR>\n"
	dot += "</TABLE> \n"
	dot += ">]; \n"
	dot += "}";

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(dot)
	reporte.Close()
	insertarReporte(id, rutaPng)
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output()
	fmt.Println( " ¡ Reporte generado exitosamente !")
}

/* Metodo que genera el reporte file, es decir, muestar el contenido de un archivo. */
func reporteFile(rutaReporte string, id string, rutaArchivo string){
	
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		mensaje +="Particion no esta montado, no es posible realizar el reporte..\n"
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		mensaje +="Disco no existe, no es posible realizar el reporte..\n"
		return
	}

	rutaLimpia := ""
	for i := 1; i < len(rutaArchivo)-1; i++ {
		rutaLimpia += string(rutaArchivo[i])
	}
	rutaLimpia += "/"

	nombreArchivoAux := "";
    for i := len(rutaArchivo)-1; i > 0; i--  {
        if string(rutaArchivo[i]) == "/" {
            break
        }
        nombreArchivoAux += string(rutaArchivo[i])
    }

	nombreArchivo := ""
	for j := len(nombreArchivoAux) - 1; j > 0; j-- {
		nombreArchivo += string(nombreArchivoAux[j])
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	inicio := super.InicioTablaInodos
	//posicion := buscarInodo(archivo, int64(inicio), rutaLimpia)
	datos := leerArchivo(archivo, super, int64(inicio), nombreArchivo)
	mensaje += "*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-\n"
	mensaje += "Archivo : " + nombreArchivo + "\n"
	mensaje += datos + "\n"
	mensaje += "*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-\n"
/* 	grafo := "digraph File {\n"
    grafo += "fontname=\"Helvetica,Arial,sans-serif\"\n"
	grafo += "node [style=filled fillcolor=\"#f8f8f8\"]\n"
    grafo += "subgraph cluster_L { \"File: [stackcollapse]\"\n"
    grafo += "[shape=box fontsize=16 label=\" Archivo : " + nombreArchivo +"\n"
    grafo += datos
    grafo += "\"] }\n"
    grafo += "}\n"

	rutaTxt, rutaPng := obtenerRutaReporte(rutaReporte)
	reporte, _ := os.Create(rutaTxt)
	reporte.WriteString(grafo)
	reporte.Close()
	
	exec.Command("dot", rutaTxt, "-Tpng", "-o", rutaPng).Output() */
	mensaje += " ¡ Reporte generado exitosamente !\n"
}

/*  Funcion que recupera el super bloque.  */
func obtenerSuperBloque(archivo *os.File, nombreParticion [20]byte, discoAux MBR) SuperBloque {

	superAux := SuperBloque{}
	inicioParticion := int64(-1)
	contenido := make([]byte, int(unsafe.Sizeof(superAux)))

	for i := 0; i < 4; i++ {
		if discoAux.Particiones[i].Nombre == nombreParticion {
			inicioParticion = discoAux.Particiones[i].Inicio
			break
		}
	}

	archivo.Seek(inicioParticion, 0)
	archivo.Read(contenido)
	buffer := bytes.NewBuffer(contenido)
	binary.Read(buffer, binary.BigEndian, &superAux)

	return superAux
}

/* Funcion que cambia la extension del archivo y genera 2 rutas para generar el grafo. */
func obtenerRutaReporte(rutaReporte string) (string, string) {

	ruta := ""
	if strings.Contains(rutaReporte, "\"") {
		ruta = strings.ReplaceAll(rutaReporte, "\"", "")
	} else {
		ruta = rutaReporte
	}

	rutaAux := strings.Split(ruta, ".")
	rutaTxt := rutaAux[0] + ".txt"
	rutaPng := rutaAux[0] + ".png"

	return rutaTxt, rutaPng
}

/*  Funcion que recorre el arreglo de byte y retorna el contenido en string */
func cadenaLimpia(cadena []byte) string {

	cadenaLimpia := ""
	for i := 0; i < len(cadena); i++ {

		if cadena[i] == 0 {
			cadenaLimpia += ""
		} else {

			cadenaLimpia += string(cadena[i])
		}
	}

	return cadenaLimpia
}


func insertarReporte(id string, ruta string){
	
	for i := 0; i < 25; i++ {
		if arregloReportes[i].IdParticion == "" {
			arregloReportes[i].IdParticion = id	
			arregloReportes[i].Ruta = ruta
			break	
		}
	}
	
}