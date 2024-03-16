package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"
)

/*
1.Struct utilizados para la administracion y creacion del sistema de archivos.

	-Struct para el super bloque.
	-Struct para la tabla de inodos.
	-Struct para el bloque de carpetas.
	-Struct para el contenido de un archivo.
	-Struct para el journaling.
*/
type SuperBloque struct {
	IdSistema            uint32
	NumeroInodos         uint32
	NumeroBloques        uint32
	NumeroBloquesLibres  uint32
	NumeroInodosLibres   uint32
	UltimaFechaMontado   [20]byte
	NumeroSistemaMontado uint32
	NumeroMagico         uint32
	TamanioInodo         uint32
	TamanioBloque        uint32
	PrimerInodoLibre     uint32
	PrimerBloqueLibre    uint32
	InicioBitMapsInodos  uint32
	InicioBitMapsBloques uint32
	InicioTablaInodos    uint32
	InicioTablaBloques   uint32
}

type TablaInodos struct {
	IdUsuario         uint32
	IdGrupo           uint32
	TamanioArchivo    uint32
	FechaLectura      [20]byte
	FechaCreacion     [20]byte
	FechaModificacion [20]byte
	Bloque            [15]int64
	Tipo              int64
	Permisos          uint32
}

type Contenido struct {
	Nombre    [25]byte
	Apuntador int64
}

type BloqueCarpeta struct {
	Contenidos [4]Contenido
}

type BloqueArchivos struct {
	Datos [150]byte
}

type Journaling struct {
	TipoOperacion string
	Tipo          byte
	Nombre        string
	Ruta          string
	Contenido     string
	Fecha         string
	Propietario   string
	Permisos      int
	Tamanio       int
}

/*  Variables gloables utilizadas para iniciar sesion y crear grupos y usuarios. */
var grupoActual = ""
var usuarioActual = ""
var contraseniaActual = ""
var sesionIniciada = false
var identificadorActual = ""

/*  Metodo que formatea una particion con el sistema de archivos EXT2. */
func crearSistemaArchivosEXT2(id string, tipoFormateo string) {

	// Obtenemos el disco montado y la particion.
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar un formateo..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar un formateo..")
		return
	}

	/*
		Obtenemos el MBR del disco.
		Recuperamos el nombre de la particion a formatear.
		Recuperamos el incio y el tamanio de la particion a formatear.
	*/
	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	inicioParticion, tamanioParticion := obtenerInicioTamanio(nombreParticion, discoAux)

	// Llenando la particion con "0"
	archivo.Seek(inicioParticion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))

	// Obtenemos el numero de estructuras para el sistema.
	numeroEstructuras := obtenerNumeroEstructuras(tamanioParticion)

	bitMapInodos := inicioParticion + int64(unsafe.Sizeof(SuperBloque{}))
	bitMapBloques := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + numeroEstructuras

	// Llenamos el super bloque con los datos.
	superBloque := SuperBloque{
		IdSistema:            2,
		NumeroInodos:         uint32(numeroEstructuras),
		NumeroBloques:        uint32(3 * numeroEstructuras),
		NumeroBloquesLibres:  uint32(3*numeroEstructuras) - 2,
		NumeroInodosLibres:   uint32(numeroEstructuras) - 2,
		NumeroSistemaMontado: 0,
		NumeroMagico:         0xEF53,
		TamanioInodo:         uint32(unsafe.Sizeof(TablaInodos{})),
		TamanioBloque:        uint32(unsafe.Sizeof(BloqueArchivos{})),
		PrimerInodoLibre:     uint32(bitMapInodos) + uint32(3*numeroEstructuras),
		PrimerBloqueLibre:    uint32(bitMapBloques) + uint32(3*numeroEstructuras),
		InicioBitMapsInodos:  uint32(bitMapInodos),
		InicioBitMapsBloques: uint32(bitMapBloques),
		InicioTablaInodos:    uint32(bitMapInodos) + uint32(numeroEstructuras) + uint32(3*numeroEstructuras),
		InicioTablaBloques:   uint32(bitMapBloques) + uint32(3*numeroEstructuras) + (uint32(numeroEstructuras) * uint32(unsafe.Sizeof(TablaInodos{}))),
	}
	copy(superBloque.UltimaFechaMontado[:], obtenerFecha())

	// Escribimos el super bloque en la particion.
	archivo.Seek(inicioParticion, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &superBloque)
	archivo.Write(buffer.Bytes())

	// Escribiendo un uno en el bitMap de inodos
	escribirBitMapInodo(archivo, uint32(numeroEstructuras), superBloque)
	// Escribiendo un uno en el bitMap de bloques
	escribirBitMapBloque(archivo, uint32(numeroEstructuras), superBloque)

	/* Creamos el Inodo para la carpeta root. */
	iNodoRoot := TablaInodos{
		IdUsuario:      1,
		IdGrupo:        1,
		TamanioArchivo: 0,
		Tipo:           int64(0),
		Permisos:       664,
	}
	copy(iNodoRoot.FechaCreacion[:], obtenerFecha())
	copy(iNodoRoot.FechaLectura[:], obtenerFecha())
	copy(iNodoRoot.FechaModificacion[:], obtenerFecha())
	for i := 0; i < 15; i++ {
		iNodoRoot.Bloque[i] = -1
	}

	posicionInodo := escribirInodo(archivo, superBloque, iNodoRoot)
	contenido := "1,G,root\n1,U,root,root,123\n"
	crearArchivo(archivo, superBloque, contenido, "users.txt", posicionInodo)

	fmt.Println("¡ Formateo del sistema EXT2 fue realizado exitosamente !")
}

func crearSistemaArchivosEXT3(id string, tipoFormateo string) {

	// Obtenemos el disco montado y la particion.
	rutaObtenida := obtenerDiscoMontado(id)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar un formateo..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	defer archivo.Close()

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar un formateo..")
		return
	}

	/*
		Obtenemos el MBR del disco.
		Recuperamos el nombre de la particion a formatear.
		Recuperamos el incio y el tamanio de la particion a formatear.
	*/
	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(id)
	inicioParticion, tamanioParticion := obtenerInicioTamanio(nombreParticion, discoAux)

	// Llenando la particion con "0"
	archivo.Seek(inicioParticion, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(0))

	// Obtenemos el numero de estructuras para el sistema.
	numeroEstructuras := obtenerNumeroEstructuras(tamanioParticion)

	bitMapInodos := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + (numeroEstructuras * int64(unsafe.Sizeof(Journaling{})))
	bitMapBloques := inicioParticion + int64(unsafe.Sizeof(SuperBloque{})) + numeroEstructuras + (numeroEstructuras * int64(unsafe.Sizeof(Journaling{})))

	// Llenamos el super bloque con los datos.
	superBloque := SuperBloque{
		IdSistema:            3,
		NumeroInodos:         uint32(numeroEstructuras),
		NumeroBloques:        uint32(3 * numeroEstructuras),
		NumeroBloquesLibres:  uint32(3*numeroEstructuras) - 2,
		NumeroInodosLibres:   uint32(numeroEstructuras) - 2,
		NumeroSistemaMontado: 0,
		NumeroMagico:         0xEF53,
		TamanioInodo:         uint32(unsafe.Sizeof(TablaInodos{})),
		TamanioBloque:        uint32(unsafe.Sizeof(BloqueArchivos{})),
		PrimerInodoLibre:     uint32(bitMapInodos) + uint32(3*numeroEstructuras),
		PrimerBloqueLibre:    uint32(bitMapBloques) + uint32(3*numeroEstructuras),
		InicioBitMapsInodos:  uint32(bitMapInodos),
		InicioBitMapsBloques: uint32(bitMapBloques),
		InicioTablaInodos:    uint32(bitMapInodos) + uint32(numeroEstructuras) + uint32(3*numeroEstructuras),
		InicioTablaBloques:   uint32(bitMapBloques) + uint32(3*numeroEstructuras) + (uint32(numeroEstructuras) * uint32(unsafe.Sizeof(TablaInodos{}))),
	}
	copy(superBloque.UltimaFechaMontado[:], obtenerFecha())

	// Escribimos el super bloque en la particion.
	archivo.Seek(inicioParticion, 0)
	buffer.Reset()
	binary.Write(buffer, binary.BigEndian, &superBloque)
	archivo.Write(buffer.Bytes())


	limite := superBloque.InicioTablaInodos
	posicionJournaling := inicioParticion + int64(unsafe.Sizeof(SuperBloque{}))
	for i := posicionJournaling; i < int64(limite); i = (i + int64(unsafe.Sizeof(Journaling{}))) {
		archivo.Seek(int64(i), 0)
		binary.Write(buffer, binary.BigEndian, &superBloque)
		archivo.Write(buffer.Bytes())

	}

	// Escribiendo un uno en el bitMap de inodos
	escribirBitMapInodo(archivo, uint32(numeroEstructuras), superBloque)
	// Escribiendo un uno en el bitMap de bloques
	escribirBitMapBloque(archivo, uint32(numeroEstructuras), superBloque)

	/* Creamos el Inodo para la carpeta root. */
	iNodoRoot := TablaInodos{
		IdUsuario:      1,
		IdGrupo:        1,
		TamanioArchivo: 0,
		Tipo:           int64(0),
		Permisos:       664,
	}
	copy(iNodoRoot.FechaCreacion[:], obtenerFecha())
	copy(iNodoRoot.FechaLectura[:], obtenerFecha())
	copy(iNodoRoot.FechaModificacion[:], obtenerFecha())
	for i := 0; i < 15; i++ {
		iNodoRoot.Bloque[i] = -1
	}

	posicionInodo := escribirInodo(archivo, superBloque, iNodoRoot)
	contenido := "1,G,root\n1,U,root,root,123\n"
	crearArchivo(archivo, superBloque, contenido, "users.txt", posicionInodo)

	fmt.Println("¡ Formateo del sistema EXT3 fue realizado exitosamente !")
}


/*  Metodo que llena de 0 de inicio a fin el mapa de bits de los inodos. */
func escribirBitMapInodo(archivo *os.File, numeroEstructuras uint32, super SuperBloque) {

	posicion := super.InicioBitMapsInodos
	contador := 0

	for {

		if contador < int(numeroEstructuras) {
			escritura := bytes.NewBuffer([]byte{})
			archivo.Seek(int64(posicion), 0)
			binary.Write(escritura, binary.BigEndian, uint8(0))
			archivo.Write(escritura.Bytes())
			contador++
			posicion++

		} else {
			break
		}
	}
}

/*  Metodo que llena de 0 de inicio a fin el mapa de bits de los bloques. */
func escribirBitMapBloque(archivo *os.File, numeroEstructuras uint32, super SuperBloque) {

	posicion := int64(super.InicioBitMapsBloques)
	contador := int64(0)
	limiteBitMap := numeroEstructuras * 3

	for {

		if int(contador) < int(limiteBitMap) {
			escritura := bytes.NewBuffer([]byte{})
			archivo.Seek(posicion, 0)
			binary.Write(escritura, binary.BigEndian, uint8(0))
			archivo.Write(escritura.Bytes())
			contador++
			posicion++
		} else {
			break
		}
	}
}

/*  Metodo que escribe en el disco un inodo. */
func escribirInodo(disco *os.File, super SuperBloque, inodo TablaInodos) int64 {

	/* Creamos una variable donde se almacenara la informacion recuperada. */
	inicioBitMapInodos := int64(super.InicioBitMapsInodos)
	aux := uint8(1)

	contador := int64(0)

	for {
		if aux == uint8(1) {

			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapInodos+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}
	/*
		1.Calculamos en donde se insertara el nuevo inodo.
		2.Nos posicionamos en ese bloque de memoria.
		3.Creamos un buffer para almacenar la info.
		4.Escribimos el inodo.
		5.Escribimos los cambios en el archivo.

		6.Nos volvemos a posicionar pero esta vez para escribir el bloque de memoria utilizado como ocupado.
		7.Escribimos 1 que significa ocupado.
		8.Creamos un buffer para almacenar la info.
		9.Escribimos el 1.
		10.Escribimos los cambios en el archivo.
	*/
	inicioNuevoInodo := int64(super.InicioTablaInodos) + (contador * int64(unsafe.Sizeof(TablaInodos{})))
	disco.Seek(inicioNuevoInodo, 0)
	bufferInodo := bytes.NewBuffer([]byte{})
	binary.Write(bufferInodo, binary.BigEndian, &inodo)
	disco.Write(bufferInodo.Bytes())

	disco.Seek(inicioBitMapInodos+contador, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	return inicioNuevoInodo
}

/*  Metodo que escribe en el disco un bloque de carpeta. */
func escribirBloqueCarpeta(disco *os.File, super SuperBloque, bloque BloqueCarpeta) int64 {

	inicioBitMapBloque := int64(super.InicioBitMapsBloques)
	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {
			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapBloque+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
			}
		} else {
			break
		}
	}

	/*
		1.Calculamos en donde se insertara el nuevo inodo.
		2.Nos posicionamos en ese bloque de memoria.
		3.Creamos un buffer para almacenar la info.
		4.Escribimos el inodo.
		5.Escribimos los cambios en el archivo.

		6.Nos volvemos a posicionar pero esta vez para escribir el bloque de memoria utilizado como ocupado.
		7.Escribimos 1 que significa ocupado.
		8.Creamos un buffer para almacenar la info.
		9.Escribimos el 1.
		10.Escribimos los cambios en el archivo.
	*/
	inicioNuevoBloque := int64(super.InicioTablaBloques) + (contador * int64(unsafe.Sizeof(BloqueCarpeta{})))
	disco.Seek(inicioNuevoBloque, 0)
	bufferBloque := bytes.NewBuffer([]byte{})
	binary.Write(bufferBloque, binary.BigEndian, &bloque)
	disco.Write(bufferBloque.Bytes())

	disco.Seek(inicioBitMapBloque+contador, 0)
	buffer := bytes.NewBuffer([]byte{})
	binary.Write(buffer, binary.BigEndian, uint8(1))
	disco.Write(buffer.Bytes())

	return inicioNuevoBloque
}

/*  Metodo que escribe en el disco un bloque de archivo. */
func escribirBloqueArchivo(disco *os.File, super SuperBloque, archivo BloqueArchivos) int64 {

	inicioBitMapBloque := int64(super.InicioBitMapsBloques)
	aux := uint8(1)
	contador := int64(0)

	for {
		if aux == uint8(1) {

			leerBitMap := make([]byte, int(unsafe.Sizeof(aux)))
			disco.Seek(inicioBitMapBloque+contador, 0)
			disco.Read(leerBitMap)
			buffer := bytes.NewBuffer(leerBitMap)
			binary.Read(buffer, binary.BigEndian, &aux)

			if aux == uint8(1) {
				contador++
				contador++
/* 				contador++
				contador++
				contador++
				contador++
				contador++ */
			}
		} else {
			break
		}
	}

	/*
		1.Calculamos en donde se insertara el nuevo inodo.
		2.Nos posicionamos en ese bloque de memoria.
		3.Creamos un buffer para almacenar la info.
		4.Escribimos el inodo.
		5.Escribimos los cambios en el archivo.

		6.Nos volvemos a posicionar pero esta vez para escribir el bloque de memoria utilizado como ocupado.
		7.Escribimos 1 que significa ocupado.
		8.Creamos un buffer para almacenar la info.
		9.Escribimos el 1.
		10.Escribimos los cambios en el archivo.
	*/
	inicioNuevoArchivo := int64(super.InicioTablaBloques) + (contador * int64(unsafe.Sizeof(BloqueArchivos{})))
	disco.Seek(inicioNuevoArchivo, 0)
	bufferBloque := bytes.NewBuffer([]byte{})
	binary.Write(bufferBloque, binary.BigEndian, &archivo)
	disco.Write(bufferBloque.Bytes())

	disco.Seek(inicioBitMapBloque+contador, 0)
	bufferuno := bytes.NewBuffer([]byte{})
	binary.Write(bufferuno, binary.BigEndian, uint8(1))
	disco.Write(bufferuno.Bytes())

	return inicioNuevoArchivo
}

/*  Metodo que crea archivos. */
func crearArchivo(archivo *os.File, super SuperBloque, cadena string, nombreArchivo string, posicionInodo int64) {

	inodo := obtenerInodo(archivo, posicionInodo)
	if inodo.Tipo == int64(0) {
		if inodo.Bloque[0] == int64(-1) {
			nuevoBloque := BloqueCarpeta{}
			copy(nuevoBloque.Contenidos[0].Nombre[:], ".")
			nuevoBloque.Contenidos[0].Apuntador = 0
			copy(nuevoBloque.Contenidos[1].Nombre[:], "..")
			nuevoBloque.Contenidos[1].Apuntador = 0
			copy(nuevoBloque.Contenidos[2].Nombre[:], "")
			nuevoBloque.Contenidos[2].Apuntador = -1
			copy(nuevoBloque.Contenidos[3].Nombre[:], "")
			nuevoBloque.Contenidos[3].Apuntador = -1
			posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
			inodo.Bloque[0] = posicionBloque
			reescribirInodo(archivo, posicionInodo, inodo)
		}

		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				posicionBloque := inodo.Bloque[i]
				bloqueCarpeta := obtenerBloqueCarpetas(archivo, posicionBloque)
				for j := 0; j < 4; j++ {
					if bloqueCarpeta.Contenidos[j].Apuntador == int64(-1) {
						nuevoInodo := TablaInodos{
							IdUsuario:      1,
							IdGrupo:        1,
							TamanioArchivo: uint32(len(cadena)),
							Tipo:           int64(1),
							Permisos:       664,
						}
						copy(nuevoInodo.FechaCreacion[:], obtenerFecha())
						copy(nuevoInodo.FechaLectura[:], obtenerFecha())
						copy(nuevoInodo.FechaModificacion[:], obtenerFecha())

						for k := 0; k < 15; k++ {
							if len(cadena) == 0 {
								nuevoInodo.Bloque[k] = -1
							} else {
								nuevoArchivo := BloqueArchivos{}

								copy(nuevoArchivo.Datos[:], cadena)
								cadena = ""
								posicionArchivo := escribirBloqueArchivo(archivo, super, nuevoArchivo)
								nuevoInodo.Bloque[k] = posicionArchivo
							}
						}

						posicionNuevoInodo := escribirInodo(archivo, super, nuevoInodo)
						bloqueCarpeta.Contenidos[j].Apuntador = posicionNuevoInodo
						copy(bloqueCarpeta.Contenidos[j].Nombre[:], nombreArchivo)
						reescribirBloque(archivo, inodo.Bloque[i], bloqueCarpeta)
						return
					}
				}

			} else {
				i--
			}
		}
	}
}

/* Metodo que reescribe un inodo */
func reescribirInodo(disco *os.File, posicionInodo int64, inodo TablaInodos) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionInodo, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &inodo)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

/* Metodo que reescribe un Bloque de carpetas */
func reescribirBloque(disco *os.File, posicionBloque int64, bloque BloqueCarpeta) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &bloque)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

/* Metodo que reescribe un Bloque de archivo */
func reescribirBloqueArchivo(disco *os.File, posicionBloque int64, bloque BloqueArchivos) {
	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//creando un buffer para almacenar la informacion requerida
	bufferEscritura := bytes.NewBuffer([]byte{})

	//codificando a binario la informacion y almacenandola en el buffer
	binary.Write(bufferEscritura, binary.BigEndian, &bloque)

	//escribiendo la informacion codificada en el archivo
	disco.Write(bufferEscritura.Bytes())
}

/* Metodo que recupera un inodo */
func obtenerInodo(disco *os.File, posicionInodo int64) TablaInodos {
	//variable que almacena el struct del avd,
	Inodo := TablaInodos{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(Inodo)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionInodo, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &Inodo)

	return Inodo
}

/* Metodo que recupera un bloque de carpetas */
func obtenerBloqueCarpetas(disco *os.File, posicionBloque int64) BloqueCarpeta {
	//variable que almacena el struct del avd,
	bloqueCarpeta := BloqueCarpeta{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(bloqueCarpeta)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &bloqueCarpeta)

	return bloqueCarpeta
}

/* Metodo que recupera un bloque de archivo */
func obtenerBloqueArchivo(disco *os.File, posicionBloque int64) BloqueArchivos {
	//variable que almacena el struct del avd,
	bloqueArchivo := BloqueArchivos{}

	//variable que almacena el contenido leido del disco
	contenido := make([]byte, int(unsafe.Sizeof(bloqueArchivo)))

	//poniendo el cursor en la posicion deseada
	disco.Seek(posicionBloque, 0)

	//obteniendo el contenido del archivo y asignandolo a la variable del contenido
	disco.Read(contenido)

	//asignando el contenido leido a un buffer para decodificar el binario
	bufferLectura := bytes.NewBuffer(contenido)

	//decodificando el contenido del buffer y asignandolo al struct
	binary.Read(bufferLectura, binary.BigEndian, &bloqueArchivo)

	return bloqueArchivo
}

/* Metodo que iniciar una sesion dada un usuario, contrase y un id para identificar a que particion utilizaremos. */
func iniciarSesion(usuario string, contrasenia string, id string) (bool, string) {

	error := ""
	encontrado := false

	if sesionIniciada {
		fmt.Println("Ya existe una sesion iniciada no puede haber 2 sesiones al mismo tiempo....")
		return encontrado, error
	}

	rutaObtenida := obtenerDiscoMontado(id)
	nombreParticion := obtenerParticionMontada(id)
	nombreParticionAux := cadenaLimpia(nombreParticion[:])

	if rutaObtenida == "" || nombreParticionAux == "" {
		fmt.Println("Particion o Disco no estan montados, no es posible realizar la accion..")
		return encontrado, error
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar la accion..")
		return encontrado, error
	}

	discoAux := obtenerMBR(archivo)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[1] == "U" {
				if letra[3] == usuario && letra[4] == contrasenia {
					usuarioActual = usuario
					contraseniaActual = contrasenia
					grupoActual = letra[2]
					identificadorActual = id
					sesionIniciada = true
					encontrado = true
				}
			}
		}
	}

	if !encontrado {
		fmt.Println("Usuario o contraseña no existe, vuelva a intentarlo....")
	} else {
		fmt.Println("¡ Bienvenido al sistema de archivos Usuario : [" + usuarioActual + "] !")
	}

	archivo.Close()

	return encontrado, error
}

/* Metodo que cierra una sesion si anteriormente se inicio una. */
func cerrarSesion() (bool, string) {

	validar := ""
	if !sesionIniciada {
		fmt.Println("¡ No existe una sesion iniciada no es posible cerrar sesion !")
		return sesionIniciada, validar
	}
	usuarioActual = ""
	contraseniaActual = ""
	grupoActual = ""
	identificadorActual = ""
	sesionIniciada = false
	fmt.Println("¡ Sesion cerrada exitosamente !")
	return sesionIniciada, validar
}

/* Metodo que crea un grupo. */
func crearGrupo(nombreGrupo string) {

	if !sesionIniciada {
		fmt.Println("Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.")
		return
	}

	if usuarioActual != "root" {
		fmt.Println("Solo el usuario [root] puede utilizar este comando.")
		return
	}

	if len(nombreGrupo) > 10 {
		fmt.Println("El nombre del grupo no puede tener mas de 10 caracteres, vuelva a intentarlo.")
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	grupoExiste := false
	idGrupo := 0

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			idGrupo, _ = strconv.Atoi(letra[0])
			if letra[1] == "G" {
				if letra[2] == nombreGrupo {
					grupoExiste = true
				}
			}
		}
	}

	if grupoExiste {
		fmt.Println("¡ Grupo ya existe, ingrese un nuevo nombre !")
		return
	}

	idGrupo++
	contenido += strconv.Itoa(int(idGrupo)) + "," + "G" + "," + nombreGrupo + "\n"
	reescribirArchivo(archivo, super, int64(posicionInicial), contenido, "users.txt")
	fmt.Println(leerArchivo(archivo, super, int64(posicionInicial), "users.txt"))
	fmt.Println("¡ Grupo creado exitosamente !")
	archivo.Close()
}

/* Metodo que elimina un grupo dado el nombre del grupo. */
func eliminarGrupo(nombreGrupo string) {

	if !sesionIniciada {
		fmt.Println("Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.")
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	concatenarContenido := ""
	grupoExiste := false

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[2] == nombreGrupo {
				letra[0] = "0"
				grupoExiste = true
			}
		}

		if letra[1] == "U" {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "," + letra[3] + "," + letra[4] + "\n"
		} else {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "\n"
		}
	}

	if !grupoExiste {
		fmt.Println("El grupo no existe, vuelva a intentarlo.")
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), concatenarContenido, "users.txt")
	fmt.Println(leerArchivo(archivo, super, int64(posicionInicial), "users.txt"))
	fmt.Println("¡ Grupo eliminado exitosamente !")
	archivo.Close()
}

/* Metodo que crea un usuario. */
func crearUsuario(usuario string, contrasenia string, grupoPertenece string) {

	if !sesionIniciada {
		fmt.Println("Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.")
		return
	}

	if usuarioActual != "root" {
		fmt.Println("Solo el usuario [root] puede utilizar este comando.")
		return
	}

	if len(contrasenia) > 10 && len(usuario) > 10 && len(grupoPertenece) > 10 {
		fmt.Println("El nombre/grupo/contraseña no pueden tener mas de 10 caracteres, vuelva a intentarlo..")
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	grupoExiste := false
	usuarioRepetido := false
	idUsuario := 0

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if !grupoExiste {
			if letra[0] != "0" {
				if letra[1] == "G" {
					if !usuarioRepetido {
						if letra[2] == grupoPertenece {
							idUsuario = idUsuario + 1
							fmt.Println("ID ", idUsuario)
							grupoExiste = true
							contenido += strconv.Itoa(int(idUsuario)) + "," + "U" + "," + grupoPertenece + "," + usuario + "," + contrasenia + "\n"
						}
					}

				} else {
					if letra[3] == usuario {
						usuarioRepetido = true
					}
				}
			}
		}
	}

	if !grupoExiste {
		fmt.Println("Grupo no existe, no es posible crear un usuario, vuelva a intentarlo.")
		return
	}

	if usuarioRepetido {
		fmt.Println("EL usuario debe ser unico, vuelva a intentarlo..")
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), contenido, "users.txt")
	fmt.Println(leerArchivo(archivo, super, int64(posicionInicial), "users.txt"))
	fmt.Println("¡ Usuario creado exitosamente !")
	archivo.Close()
}

/* Metodo que elimina un usuario dado el nombre del usuario. */
func eliminarUsuario(nombreUsuario string) {

	if !sesionIniciada {
		fmt.Println("Debe existir una sesion iniciada para crear un grupo, vuelva  intentarlo.")
		return
	}

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte..")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte..")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos

	contenido := leerArchivo(archivo, super, int64(posicionInicial), "users.txt")
	separarContenido := strings.Split(contenido, "\n")
	concatenarContenido := ""
	grupoExiste := false

	for i := 0; i < len(separarContenido)-1; i++ {
		letra := strings.Split(separarContenido[i], ",")
		if letra[0] != "0" {
			if letra[1] == "U" {
				if letra[3] == nombreUsuario {
					letra[0] = "0"
					grupoExiste = true
				}
			}
		}

		if letra[1] == "U" {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "," + letra[3] + "," + letra[4] + "\n"
		} else {
			concatenarContenido += letra[0] + "," + letra[1] + "," + letra[2] + "\n"
		}
	}

	if !grupoExiste {
		fmt.Println("El usuario no existe, vuelva a intentarlo.")
		return
	}

	reescribirArchivo(archivo, super, int64(posicionInicial), concatenarContenido, "users.txt")
	fmt.Println(leerArchivo(archivo, super, int64(posicionInicial), "users.txt"))
	fmt.Println("¡ Usuario eliminado exitosamente !")
	archivo.Close()
}

/* Funcion que retorna el contenido de un archivo. */
func leerArchivo(archivo *os.File, super SuperBloque, posicion int64, nombreArchivo string) string {

	contenido := ""
	inodo := obtenerInodo(archivo, posicion)
	if inodo.Tipo == int64(0) {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if carpeta.Contenidos[j].Apuntador != int64(-1) && nombreLimpio == nombreArchivo {
						inodoAux := obtenerInodo(archivo, carpeta.Contenidos[j].Apuntador)
						for k := 0; k < 15; k++ {
							if inodoAux.Bloque[k] != int64(-1) {
								archivo := obtenerBloqueArchivo(archivo, inodoAux.Bloque[k])
								contenido += cadenaLimpia(archivo.Datos[:])
							}
						}
						if inodoAux.Bloque[14] == -1 {
							return contenido
						}
					}
				}
			}
		}
	}
	return "null"
}

/* Funcion que analiza los parametros correspondiente para la creacion de carpetas */
func crearCarpeta(ruta string, padre bool) {

	rutaObtenida := obtenerDiscoMontado(identificadorActual)

	if rutaObtenida == "" {
		fmt.Println("Particion no esta montado, no es posible realizar el reporte.")
		return
	}

	// Abrimos el archivo y verificamos su existencia.
	archivo, _ := os.OpenFile(rutaObtenida, os.O_RDWR, 0644)

	if archivo == nil {
		fmt.Println("Disco no existe, no es posible realizar el reporte.")
		return
	}

	discoAux := obtenerMBR(archivo)
	nombreParticion := obtenerParticionMontada(identificadorActual)
	inicioPart, _ := obtenerInicioTamanio(nombreParticion, discoAux)
	super := obtenerSuperBloque(archivo, nombreParticion, discoAux)
	posicionInicial := super.InicioTablaInodos
	contador := 0

	rutaLimpia := ""
	for i := 1; i < len(ruta); i++ {
		rutaLimpia += string(ruta[i])
	}

	ultimaCarpetaAlreves := ""
	for j := len(ruta) - 1; j >= 0; j-- {
		if string(ruta[j]) == "/" {
			break
		}
		contador++
		ultimaCarpetaAlreves += string(ruta[j])
	}
	
	ultimaCarpeta := ""
	for j := len(ultimaCarpetaAlreves) - 1; j >= 0; j-- {
		ultimaCarpeta += string(ultimaCarpetaAlreves[j])
	}
	
	nombrePadreAux := ""
	for k := len(ruta) - contador; k >= 0; k-- {
		nombrePadreAux += string(ruta[k])
	}
	
	rutaPadre := ""
	for l := len(nombrePadreAux) - 1; l > 0; l-- {
		rutaPadre += string(nombrePadreAux[l])
	}

	if padre {
		mkdirCOmplemento(archivo, super, rutaLimpia+"/", int64(posicionInicial), inicioPart)
	} else {

		mismaCarpeta := buscarCarpetaArchivo(archivo, int64(posicionInicial), rutaLimpia+"/", ultimaCarpeta)
		carpetaPadre := buscarInodo(archivo, int64(posicionInicial), rutaPadre)

		if mismaCarpeta == 1 {
			fmt.Println("¡ No es posible crear la misma carpetas !")
			return
		} else if carpetaPadre == 1 {
			fmt.Println("¡ No es posible crear las carpetas, no existen carpetas padres !")
			return
		} else {
			mkdirCOmplemento(archivo, super, rutaLimpia+"/", int64(posicionInicial), inicioPart)
		}
	}
}

/* 
	Metodo que de acuerdo a la ruta enviada busca la posicion de la carpeta si la encuentra retorna un valor de lo contrario sera
	cero y de ser asi entonces quiere decir que hay que crear una carpeta nueva.
 */
var pos int = 0
func mkdirCOmplemento(archivo *os.File, super SuperBloque, ruta string, posicionTablaInodos int64, inicioParticion int64) {

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			pos = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := pos; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	posicion := buscarInodo(archivo, posicionTablaInodos, nombre+"/")

	if len(nombre) > 0 {
		if posicion == 0 {
			crearDirectorio(archivo, super, nombre, posicionTablaInodos)
			otraPosicion := buscarInodo(archivo, posicionTablaInodos, nombre+"/")
			if len(nombre) > 0 {
				mkdirCOmplemento(archivo, super, retoRuta, otraPosicion, inicioParticion)
			}
			return
		} else {
			mkdirCOmplemento(archivo, super, retoRuta, posicion, inicioParticion)
		}
	}
}

/* FUncion que busca la posicion de un archivo o carpeta dada su ruta, retorna cero si no la encuentra. */
var posicionEFE int = 0
func buscarInodo(archivo *os.File, posicion int64, ruta string) int64 {

	inodo := obtenerInodo(archivo, posicion)

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			posicionEFE = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := posicionEFE; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	if inodo.Tipo == 0 {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != -1 {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if nombreLimpio == nombre {
						if len(retoRuta) == 0 {
							return carpeta.Contenidos[j].Apuntador
						} else {

							buscarInodo(archivo, carpeta.Contenidos[j].Apuntador, retoRuta)
						}
					}
				}
			}
		}
		if inodo.Bloque[14] == -1 {
			return 0
		}
	}
	return 0
}

/* Funcion que crea el inodo y los bloques correspondiente para la creacion de una carpeta. */
func crearDirectorio(archivo *os.File, super SuperBloque, nombreCarpeta string, posicionInodo int64) {

	inodo := obtenerInodo(archivo, posicionInodo)
	if inodo.Tipo == int64(0) {
		if inodo.Bloque[0] == int64(-1) {
			nuevoBloque := BloqueCarpeta{}
			copy(nuevoBloque.Contenidos[0].Nombre[:], ".")
			nuevoBloque.Contenidos[0].Apuntador = 0
			copy(nuevoBloque.Contenidos[1].Nombre[:], "..")
			nuevoBloque.Contenidos[1].Apuntador = 0
			copy(nuevoBloque.Contenidos[2].Nombre[:], "")
			nuevoBloque.Contenidos[2].Apuntador = -1
			copy(nuevoBloque.Contenidos[3].Nombre[:], "")
			nuevoBloque.Contenidos[3].Apuntador = -1
			posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
			inodo.Bloque[0] = posicionBloque
			reescribirInodo(archivo, posicionInodo, inodo)
		}

		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				posicionBloque := inodo.Bloque[i]
				bloqueCarpeta := obtenerBloqueCarpetas(archivo, posicionBloque)
				for j := 0; j < 4; j++ {
					if bloqueCarpeta.Contenidos[j].Apuntador == int64(-1) {
						nuevoInodo := TablaInodos{
							IdUsuario:      1,
							IdGrupo:        1,
							TamanioArchivo: 0,
							Tipo:           int64(0),
							Permisos:       664,
						}
						copy(nuevoInodo.FechaCreacion[:], obtenerFecha())
						copy(nuevoInodo.FechaLectura[:], obtenerFecha())
						copy(nuevoInodo.FechaModificacion[:], obtenerFecha())
						for k := 0; k < 15; k++ {
							nuevoInodo.Bloque[k] = -1
						}
						posicionNuevoInodo := escribirInodo(archivo, super, nuevoInodo)
						bloqueCarpeta.Contenidos[j].Apuntador = posicionNuevoInodo
						copy(bloqueCarpeta.Contenidos[j].Nombre[:], nombreCarpeta)
						reescribirBloque(archivo, inodo.Bloque[i], bloqueCarpeta)
						return
					}
				}

			} else {
				nuevoBloque := BloqueCarpeta{}
				copy(nuevoBloque.Contenidos[0].Nombre[:], "")
				nuevoBloque.Contenidos[0].Apuntador = -1
				copy(nuevoBloque.Contenidos[1].Nombre[:], "")
				nuevoBloque.Contenidos[1].Apuntador = -1
				copy(nuevoBloque.Contenidos[2].Nombre[:], "")
				nuevoBloque.Contenidos[2].Apuntador = -1
				copy(nuevoBloque.Contenidos[3].Nombre[:], "")
				nuevoBloque.Contenidos[3].Apuntador = -1
				posicionBloque := escribirBloqueCarpeta(archivo, super, nuevoBloque)
				inodo.Bloque[i] = posicionBloque
				reescribirInodo(archivo, posicionInodo, inodo)
				i--
			}
		}
	}
}

/* Busca una carpeta o archivo dada una ruta, si la encuentra retornara 1. */
var posicionBuscar int = 0
func buscarCarpetaArchivo(archivo *os.File, posicion int64, ruta string, nombreCarpeta string) int64 {

	inodo := obtenerInodo(archivo, posicion)

	nombre := ""
	for i := 0; i < len(ruta); i++ {
		if string(ruta[i]) == "/" {
			posicionBuscar = int(i + 1)
			break
		}
		nombre += string(ruta[i])
	}

	retoRuta := ""
	for j := posicionBuscar; j < len(ruta); j++ {
		retoRuta += string(ruta[j])
	}

	if inodo.Tipo == 0 {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != -1 {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if nombreLimpio == nombre {
						if len(retoRuta) == 0 {
							return 1
						} else {
							return buscarCarpetaArchivo(archivo, carpeta.Contenidos[j].Apuntador, retoRuta, nombreCarpeta)
						}
					}
				}
			}
		}
		if inodo.Bloque[14] == -1 {
			return 0
		}
	}
	return 0
}

/* Metodo que reescribe el archivo de users.txt  */
func reescribirArchivo(archivo *os.File, super SuperBloque, posicion int64, contenido string, nombreArchivo string) {

	inodo := obtenerInodo(archivo, posicion)
	if inodo.Tipo == int64(0) {
		for i := 0; i < 15; i++ {
			if inodo.Bloque[i] != int64(-1) {
				carpeta := obtenerBloqueCarpetas(archivo, inodo.Bloque[i])
				for j := 0; j < 4; j++ {
					nombreLimpio := cadenaLimpia(carpeta.Contenidos[j].Nombre[:])
					if carpeta.Contenidos[j].Apuntador != int64(-1) && nombreLimpio == nombreArchivo {
						inodoAux := obtenerInodo(archivo, carpeta.Contenidos[j].Apuntador)
						for k := 0; k < 15; k++ {
							if inodoAux.Bloque[k] != int64(-1) {
								archivoUsr := obtenerBloqueArchivo(archivo, inodoAux.Bloque[k])
								copy(archivoUsr.Datos[:], contenido)
								reescribirBloqueArchivo(archivo, inodoAux.Bloque[k], archivoUsr)
							}
						}
						reescribirInodo(archivo, carpeta.Contenidos[j].Apuntador, inodoAux)
					}
				}
			}
		}
	}
}

/* Funcion que obtiene el numero de bloques y carpetas que contendra nuestro sistema de archivos. */
func obtenerNumeroEstructuras(tamanioParticion int64) int64 {

	tamanioSuperBloque := unsafe.Sizeof(SuperBloque{})
	tamanioInodo := unsafe.Sizeof(TablaInodos{})
	tamanioBloque := unsafe.Sizeof(BloqueArchivos{})

	numerador := tamanioParticion - int64(tamanioSuperBloque)
	denominador := 4 + tamanioInodo + (3 * tamanioBloque)

	return numerador / int64(denominador)
}
