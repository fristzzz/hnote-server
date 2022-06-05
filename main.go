package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/thedevsaddam/renderer"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var db *mgo.Database
var rnd renderer.Render

const (
	hostName       = "localhost:27017"
	dbName         = "hnote"
	collectionName = "notes"
	port           = ":443"
	testPort       = ":8080"
	certFile       = "cert.pem"
	keyFile        = "key.pem"
)

type (
	// Note json 数据
	Note struct {
		ID           string    `json:"id"`
		Title        string    `json:"title"`
		CreateDate   time.Time `json:"create_date"`
		LastEditDate time.Time `json:"last_edit_time"`
		Content      string    `json:"content"`
	}

	// Note bson 数据
	NoteModel struct {
		ID           bson.ObjectId `bson:"_id, omitempty"`
		Title        string        `bson:"title"`
		CreateDate   time.Time     `bson:"create_date"`
		LastEditDate time.Time     `bson:"last_edit_time"`
		Content      string        `bson:"content"`
	}
)

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// 初始化， 为rnd和db赋值
func init() {
	rnd = *renderer.New()
	sess, err := mgo.Dial(hostName)
	checkErr(err)
	sess.SetMode(mgo.Monotonic, true)
	db = sess.DB(dbName)
}

func main() {
	stopServer := make(chan os.Signal)
	signal.Notify(stopServer, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// chi 路由
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	// /note路径处理笔记的增删改查
	r.Get("/note/get", getNote)
	r.Post("/note/create", createNote)
	r.Delete("/note/delete", deleteNote)
	r.Put("/note/update", updateNote)
	r.Get("/", homeHandler)

	srv := &http.Server{
		//Addr:         port,
		Addr:         testPort,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on port: %s\n", srv.Addr)

		// 测试用http
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("listen: %s\n", err)
			return
		}

		// 需要将cert.pem 和 key.pem 文件放在主文件同一目录下
		//		if err := srv.ListenAndServeTLS(certFile, keyFile); err != nil {
		//			log.Printf("error occurs running server: %s\n", err)
		//			return
		//		}
	}()

	<-stopServer
	log.Printf("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// 后续处理

		cancel()
	}()
	srv.Shutdown(ctx)
}

func getNote(w http.ResponseWriter, r *http.Request) {
}

func createNote(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("creating"))
	n := &Note{}
	if err := json.NewDecoder(r.Body).Decode(n); err != nil {
		log.Printf("erroer decoding json: %+v", err)
		w.Write([]byte("error decode json"))
		//rnd.JSON(w, http.StatusProcessing, err)
		return
	}

	nm := &NoteModel{
		ID:           bson.NewObjectId(),
		CreateDate:   n.CreateDate,
		LastEditDate: n.LastEditDate,
		Content:      n.Content,
		Title:        n.Title,
	}

	if err := db.C(collectionName).Insert(nm); err != nil {
		rnd.JSON(w, http.StatusBadRequest, renderer.M{
			"message": "failed to create note",
			"err":     err,
		})
		return
	}

	rnd.JSON(w, http.StatusCreated, renderer.M{
		"message": "note create succeed",
		"note_id": nm.ID.Hex(),
	})
}

func updateNote(w http.ResponseWriter, r *http.Request) {
}

func deleteNote(w http.ResponseWriter, r *http.Request) {

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello! this is hnote"))
}
