/**
 *Author: JOHAN SEBASTIAN FUENTES ORTEGA
 * e-mail: Johan.fuentes01@uceva.edu.co
 * Date:14/02/2022
**/
package main
import(
	"log"
	"net"
	"fmt"
	"encoding/json"
	"path"
	"io/ioutil"
	//"math" <-La uso cuando quiero enviar un archivo de gran tamaño a otro cliente usando 
)

type Peticiones struct {
	Peticion int
	Archivo []byte
	FormatoArch string
	Suscripcion []int
}
/*
--Visualizador es una variable global que almacenará los canales recibidos por el servidor
--para evitar que el cliente siempre los esté pidiendo
--suscrito es la variable global que almacena los canales a los que se suscribió el cliente
*/
var visualizador, suscrito []int

func main(){
	//Realizo la conexión con el servidor indicando el método con su dirección y puerto
	c, err := net.Dial("tcp","127.0.0.1:20")	
	check("\t-----Bienvenido-----",err)
	opcion :=-1
	defer c.Close()
	for {
		fmt.Println("\n \tMenu principal")
		fmt.Println("Por favor digite la opción a escoger:\n"+
					"1: Suscribirse a un canal\n"+
					"2: Visualizar canales suscritos\n"+
					"3: Enviar\n"+
					"4: Recibir\n"+
					"0: Salir\n")
		fmt.Print(">>")
		fmt.Scan(&opcion)
		switch opcion {
		case 1:
			suscribirseCanal(c)
		case 2:
			visualizarCanales()
		case 3:
			//Si el cliente no se ha suscrito al menos a un canal no lo deja enviar archivos
			if visualizador!=nil{
				p:=escogerArchivoCanal()
				if p.Peticion==3{
					enviarDatos(c,p,"Datos Enviados")}
			}else{
				fmt.Println("Debe suscribirse al menos a un canal")
			}
		case 4:
			//Si el cliente no se ha suscrito al menos a un canal no lo deja recibir archivos
			if visualizador!=nil{
				//Solicito la recepción de un archivo
				p:=Peticiones{Peticion:4,Suscripcion:suscrito}
				enviarDatos(c,p,"Solicitud completada")
				p=leerConeccion(c)
				if p.Peticion!=-1{
					//Cuando llega el archivo le asigno el nombre y el formato
					//el * segun la documentación de Go genera numeros aleatorios cuando se genere como "archivo temporal"
					ruta:="Archivo*"+p.FormatoArch
					//con tempfile le asigno la ubicación y nombre del archivo, en este caso el . indica que lo
					//guardará en la misma carpeta del ejecutable, asignar la dirección donde lo quieres guardar
					tmpfile, err := ioutil.TempFile(".", ruta)
					if err != nil {
						log.Println(err)
					}
					fmt.Printf("El nombre del archivo es: %s", tmpfile.Name())
					//escribo el archivo
					if _, err := tmpfile.Write(p.Archivo); err != nil {
						tmpfile.Close()
						log.Println(err)
					}
					if err := tmpfile.Close(); err != nil {
						log.Println(err)
					}
				}else{
					//Muestra la respuesta del servidor en caso que un cliente no haya enviado primeramente un archivo
					fmt.Printf("%s\n",p.Archivo)
				}
			}else{
				fmt.Println("Debe suscribirse al menos a un canal")
			}
		case 0:
			p:=Peticiones{Peticion:0}
			enviarDatos(c,p,"Gracias por usar este programa para transmitir datos!")
			return
		default:
			fmt.Println("Opción no valida, digite un valor válido por favor.")
		}
	}

}

func suscribirseCanal(c net.Conn){
	var canales int
	if visualizador==nil{
		//solicito los canales al servidor
		p:=Peticiones{Peticion:1}
		enviarDatos(c,p,"")
		p=leerConeccion(c)

		visualizador = p.Suscripcion
		//Asigno a suscrito el mismo tamaño del visualizador
		suscrito = make([]int,len(visualizador))
	}
	estado:=visualizarCanales()
	if !estado{
		for{
			//Suscribo a los canales
			fmt.Println("Digite el numero del canal a suscribirse:")
			fmt.Scan(&canales)
			if canales>0 && canales<(len(visualizador)+1){
				suscrito[canales-1]=canales
				break
			}
		}
	}
}

func visualizarCanales() bool{
	estado:=false
	if visualizador!=nil {
			var aux,aux1 string
			fmt.Println("\nCanales a suscribirse")
			for i,v := range visualizador{
				if v != suscrito[i]{
					//concateno el string mostrado en pantalla para mostrar los canales a los que no
					//está suscrito el cliente
					aux+= fmt.Sprintf("%d) Canal %d \t",i+1,v)
				}else{
					aux1+= fmt.Sprintf("-Canal %d \t",i+1)
				}
			}
			if aux1 ==""{ aux1="0 Canales suscritos"}
			if aux ==""{ 
					aux="\nYa se encuentra suscrito a todos los canales"
					estado=true
				}
			fmt.Println(aux,"\n Canales suscritos: \n", aux1)
			return estado
	}else{
		fmt.Println("Usted no está suscrito a ningun Canal")
		return estado
	}
}

func escogerArchivoCanal() Peticiones{
	p := Peticiones{}
	var ruta string
	fmt.Println("Ingrese la ruta del archivo")
	fmt.Scan(&ruta)
	//Obtengo el formato del archivo
	formato:=path.Ext(ruta)
	//busco el archivo para obtenerlo si no lo obtengo retorno una estructura peticion nil
	archivo,err:=ioutil.ReadFile(ruta)
	if err != nil {
		fmt.Println("\n No se encontró el archivo")
		return p
	}
	fmt.Println("Ingrese el canal por el que lo quiere enviar")
	for i := 0; i < len(suscrito); i++ {
		if suscrito[i]!=0 {
			//Muestro en pantalla los canales a los que esta suscrito el cliente
			aux:= fmt.Sprintf("%d) Canal  %d\t",i+1,suscrito[i])
			fmt.Print(aux)
		}
	}
	p.Peticion =3
	p.FormatoArch = formato
	p.Archivo = archivo
	opcion :=-1
	for{	
		fmt.Print("\n ingrese 0 para no escoger más. \n >>")
		fmt.Scan(&opcion)
		if opcion == 0{
			break
		}
		//agrego los canales por los que enviaré el archivo
		p.Suscripcion = append(p.Suscripcion,opcion)
	}
	return p
}

func enviarDatos(c net.Conn, p Peticiones, sms string){
	//encripto la estructura petición y envío al servidor
	enc,err := json.Marshal(p)
	if err != nil {
		log.Fatal("error:",err)
	}
	c.Write(enc)
	check(sms,err)
}

func leerConeccion(c net.Conn) Peticiones{
	/*--Si lo que se busca compartir son archivos de poco menos de 2Gb usar el
	  --el metodo de math.MaxInt32 en el tamaño del buffer*/
	buf := make([]byte, 0, 20700000)
	//me retorna un entero que me dará la capacidad del bufer que envía el cliente
	r,_ := c.Read(buf[len(buf):cap(buf)])
	//igualo el buf al tamaño de los datos enviados por el cliente. 
	//Internamente Go asigna al buf los valores respectivos de los datos enviados por el cliente
	buf = buf[:len(buf)+r]
	p:=Peticiones{}
	//desencripto los valores que llegan y lo asigno a la variable de tipo Peticion
	json.Unmarshal(buf,&p)
	return p
}
//Funcion para verificar si hubo o no erroes
func check(sms string, err error){
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sms)
}