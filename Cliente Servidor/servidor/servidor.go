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
	//"math" <-La uso cuando quiero enviar un archivo de gran tamaño a otro cliente usando 
)

type ClienteJob struct{
	Estado string
	Conn net.Conn
	Datos Person
}

type Person struct {
	Peticion int
	Archivo []byte
	FormatoArch string
	Suscripcion []int
} 
//Funcion para verificar si hubo o no erroes
func check(sms string, err error){
	if err!= nil{
		log.Fatal(err)
	}
	fmt.Printf("%s\n",sms)
}

var listCanales = []int{1,2,3,4}//los canales disponibles disponibles a los que se puede suscribir un cliente
var contar=0 					//Contador que me indicará el tamaño del canal
var auxP = Person{}

func main(){
	//Creo el canal clientJobs quien administrará el envío y recepción de archivos de los clientes
	clientJobs := make(chan ClienteJob)
	go generarRespuesta(clientJobs)//Hilo que administrará el envío de los archivos
	
	//Habilito la dirección local y el puerto para la conexión de los clientes con el servidor
	v:= &net.TCPAddr{IP: net.IPv4(127,0,0,1), Port:20}
	//le estipulo que el método de conexión será por TCP a la dirección
	lectura,err := net.ListenTCP("tcp",v)
	check("Servidor Iniciado", err)
	defer lectura.Close()
	for{
		conn,err := lectura.Accept()//Aceptamos la conexión del cliente
		check("Conexion aceptada",err)
		go func(){
			/*--Si lo que se busca compartir son archivos de poco menos de 2Gb usar el
			  --el metodo de math.MaxInt32 en el tamaño del buffer*/
			buf := make([]byte, 0, 20700000)
			for{
				//me retorna un entero que me dará la capacidad del bufer que envía el cliente
				r,_ := conn.Read(buf[len(buf):cap(buf)])
				//igualo el buf al tamaño de los datos enviados por el cliente. 
				//Internamente Go asigna al buf los valores respectivos de los datos enviados por el cliente
				buf = buf[:len(buf)+r]
				p:=Person{}
				//Debido a que lo que llega en el buf es la encriptación de un archivo json,
				//Los desencripto con Unmarshal y lo asigno a la estructura Person
				json.Unmarshal(buf,&p)
				switch p.Peticion {
				case 1:
					//Si lo que pide el cliente son la lista de los canales, se las encripto y las retorno 
					p.Suscripcion =listCanales
					enc,_ := json.Marshal(p) 
					conn.Write(enc)
				case 3:
					//aquí agrego al canal los clientes que envían archivos
					contar++
					clientJobs <- ClienteJob{"enviar",conn,p}
				case 4:
					//aquí agrego al canal los clientes que envían archivos
					contar++
					clientJobs <- ClienteJob{"recibir",conn,p}
				case 0:
					//Cierro la conexión del cliente cuando la solicita
					log.Println("Cliente desconectado")
					conn.Close()
					return	
				}
				//fmt.Println("contar:",contar)
				//reasigno el buf su valor inicial
				buf = make([]byte, 0, 20700000)
			}
		}()
	}
}

func generarRespuesta(clientJobs chan ClienteJob){
	for{ 
		for i:=0;i<contar;i++{
			res:= <- clientJobs
			if res.Estado =="enviar"{
				//obtengo el archivo, los canales suscritos y el formato del archivo del cliente que envia a otro
				auxP.Archivo = res.Datos.Archivo
				auxP.Suscripcion = res.Datos.Suscripcion
				auxP.FormatoArch = res.Datos.FormatoArch
			}
			if res.Estado =="recibir"{
				if auxP.Archivo!=nil{
					canalAux:=res.Datos.Suscripcion //obtengo los canales de los clientes que reciben
					sacar:=false
					for _,v := range auxP.Suscripcion{
						for _,v2 :=range canalAux{
							if v==v2{
								//en caso que al menos uno de los canales del cliente que envía y 
								//el cliente que recibe son los mismos
								//asigno el archivo y su formato al receptor, y los encripto para enviar
								res.Datos.Archivo = auxP.Archivo
								res.Datos.FormatoArch = auxP.FormatoArch
								enc,errr:=json.Marshal(res.Datos)
								if errr != nil {
									log.Println(errr)
								}
								res.Conn.Write(enc)
								sacar=true
								break
							}
						}
						if sacar{break}
					}
					
				}else{
					//Respuesta que devuelvo a los receptores que solicitan su petición antes que el 
					//cliente que envía envíe el archivo
					res.Datos.Archivo = []byte("Primero deben enviarle el archivo antes de solicitarlo")
					res.Datos.Peticion=-1
					enc,errr:=json.Marshal(res.Datos)
					if errr != nil {
						log.Println(errr)
					}
					res.Conn.Write(enc)
				}
			}
			/*
			--Cuando lo que se recibe son archivos muy grandes, limito el indice de peticiones a 5
			--y reinicio los canales para que la capacidad de la memoria no se sobrepase
			if contar==5{  
				auxP=Person{}
				contar=0
			}
			*/
		}
	}
	
}