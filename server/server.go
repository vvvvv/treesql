package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/pzhin/go-sophia"
)

const (
	KeyTemplate   = "key%v"
	ValueTemplate = "value%v"

	DBPath       = "sophia" // TODO: parse from command-line flag
	DBName       = "test"
	RecordsCount = 500000

	RecordsCountBench = 5000000
)

const Port = "6000" // TODO: parse from command-line flag

func main() {
	fmt.Println("TreeSQL server")

	fmt.Println("new env")
	env := newEnvironment()
	fmt.Println("new db")
	db := newDatabase(env)

	listeningSock, _ := net.Listen("tcp", ":"+Port)
	fmt.Printf("listening on port %s\n", Port)

	connectionID := 0
	for {
		conn, _ := listeningSock.Accept()
		go handleConnection(conn, connectionID, env, db)
		connectionID++
	}
}

func handleConnection(conn net.Conn, connID int, env *sophia.Environment, db *sophia.Database) {
	fmt.Printf("connection id %d from %s\n", connID, conn.RemoteAddr())
	for {
		// will listen for message to process ending in newline (\n)
		message, err := bufio.NewReader(conn).ReadString('\n')

		if err != nil {
			fmt.Printf("conn id %d terminated: %v\n", connID, err)
			return
		}

		// output message received
		fmt.Print("Message Received:", string(message))
		// sample process for string received
		newmessage := strings.ToUpper(message)
		// send new string back to client
		conn.Write([]byte(newmessage + "\n"))
	}
}

func newEnvironment() *sophia.Environment {
	env, _ := sophia.NewEnvironment()
	return env
}

func newDatabase(env *sophia.Environment) *sophia.Database {
	env.Set("sophia.path", DBPath)

	schema := &sophia.Schema{}
	schema.AddKey("key", sophia.FieldTypeString)
	schema.AddValue("value", sophia.FieldTypeString)

	db, _ := env.NewDatabase(&sophia.DatabaseConfig{
		Name:   DBName,
		Schema: schema,
	})
	env.Open()
	return db
}

func doConcurrentTx(env *sophia.Environment, db *sophia.Database) {
	fmt.Println("starting initial writes")
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		doc.Set("key", fmt.Sprintf(KeyTemplate, i))
		doc.Set("value", fmt.Sprintf(ValueTemplate, i))

		db.Set(doc)
		doc.Free()
	}
	fmt.Println("finished initial writes")

	tx1, _ := env.BeginTx()
	tx2, _ := env.BeginTx()

	go func() {
		fmt.Println("starting tx1 writes")
		for i := 0; i < RecordsCount; i++ {
			doc := db.Document()
			value := fmt.Sprintf(ValueTemplate, i+1)
			doc.Set("key", fmt.Sprintf(KeyTemplate, i))
			doc.Set("value", value)

			tx1.Set(doc)
			doc.Free()
		}
		tx1.Commit()

		fmt.Println("finished tx1 writes")
	}()

	go func() {
		fmt.Println("starting tx2 writes")
		for i := 0; i < RecordsCount; i++ {
			doc := db.Document()
			doc.Set("key", fmt.Sprintf(KeyTemplate, i))
			value := fmt.Sprintf(ValueTemplate, i+2)
			doc.Set("value", value)

			tx2.Set(doc)
			doc.Free()
		}
		tx2.Commit()

		fmt.Println("finished tx2 writes")
	}()

	fmt.Println("sleeping for 30s")
	time.Sleep(time.Duration(30) * time.Second)

	fmt.Println("reading")
	var size int
	for i := 0; i < RecordsCount; i++ {
		doc := db.Document()
		doc.Set("key", fmt.Sprintf(KeyTemplate, i))

		d, _ := db.Get(doc)
		value := d.GetString("value", &size)
		fmt.Printf("read %s\n", value)
		doc.Free()
		d.Free()
		d.Destroy()
	}
	fmt.Println("done reading")
}
