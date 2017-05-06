package main

import (
	"crypto/md5"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
	//Import this by exec in CLI: `go get -u github.com/TexaProject/texalib`
	"github.com/TexaProject/texajson"
	"github.com/TexaProject/texalib"
)

// AIName exports form value from /welcome globally
var AIName string

// IntName exports form value from /texa globally
var IntName string

func rootHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/welcome", 301)
}

func texaHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		// fmt.Printf("%+v\n", r.Form)

		fmt.Println("--INTERROGATION FORM DATA--")
		IntName = r.Form.Get("IntName")
		QSA := r.Form.Get("scoreArray")
		SlabName := r.Form.Get("SlabName")
		slabSequence := r.Form.Get("slabSequence")

		fmt.Println("###", AIName)
		fmt.Println("###", IntName)
		fmt.Println("###", QSA)
		fmt.Println("###", SlabName)
		fmt.Println("###", slabSequence)

		// LOGIC
		re := regexp.MustCompile("[0-1]+")
		array := re.FindAllString(QSA, -1)

		SlabNameArray := regexp.MustCompile("[,]").Split(SlabName, -1)
		slabSeqArray := regexp.MustCompile("[,]").Split(slabSequence, -1)

		fmt.Println("###Resulting Array:")
		for x := range array {
			fmt.Println(array[x])
		}

		fmt.Println("###SlabNameArray: ")
		fmt.Println(SlabNameArray)

		fmt.Println("###slabSeqArray: ")
		fmt.Println(slabSeqArray)

		ArtiQSA := texalib.Convert(array)
		fmt.Println("###ArtiQSA:")
		fmt.Println(ArtiQSA)

		HumanQSA := texalib.SetHumanQSA(ArtiQSA)
		fmt.Println("###HumanQSA:")
		fmt.Println(HumanQSA)

		TSA := texalib.GetTransactionSeries(ArtiQSA, HumanQSA)
		fmt.Println("###TSA:")
		fmt.Println(TSA)

		ArtiMts := texalib.GetMeanTestScore(ArtiQSA)
		HumanMts := texalib.GetMeanTestScore(HumanQSA)

		fmt.Println("###ArtiMts: ", ArtiMts)
		fmt.Println("###HumanMts: ", HumanMts)

		PageArray := texajson.GetPages()
		fmt.Println("###PageArray")
		fmt.Println(PageArray)
		for _, p := range PageArray {
			fmt.Println(p)
		}

		newPage := texajson.ConvtoPage(AIName, IntName, ArtiMts, HumanMts)

		PageArray = texajson.AddtoPageArray(newPage, PageArray)
		fmt.Println("###AddedPageArray")
		fmt.Println(PageArray)

		JsonPageArray := texajson.ToJson(PageArray)
		fmt.Println("###jsonPageArray:")
		fmt.Println(JsonPageArray)

		////
		fmt.Println("### SLAB LOGIC")

		slabPageArray := texajson.GetSlabPages()
		fmt.Println("###slabPageArray")
		fmt.Println(slabPageArray)

		slabPages := texajson.ConvtoSlabPage(ArtiQSA, SlabNameArray, slabSeqArray)
		fmt.Println("###slabPages")
		fmt.Println(slabPages)
		for z := 0; z < len(slabPages); z++ {
			slabPageArray = texajson.AddtoSlabPageArray(slabPages[z], slabPageArray)
		}
		fmt.Println("###finalslabPageArray")
		fmt.Println(slabPageArray)

		JsonSlabPageArray := texajson.SlabToJson(slabPageArray)
		fmt.Println("###JsonSlabPageArray: ")
		fmt.Println(JsonSlabPageArray)

		////
		fmt.Println("### CAT LOGIC")

		CatPageArray := texajson.GetCatPages()
		fmt.Println("###CatPageArray")
		fmt.Println(CatPageArray)

		CatPages := texajson.ConvtoCatPage(AIName, slabPageArray, SlabNameArray)
		fmt.Println("###CatPages")
		fmt.Println(CatPages)
		CatPageArray = texajson.AddtoCatPageArray(CatPages, CatPageArray)

		// for z := 0; z < len(CatPages); z++ {
		// 	CatPageArray = texajson.AddtoCatPageArray(CatPages[z], CatPageArray)
		// }
		fmt.Println("###finalCatPageArray")
		fmt.Println(CatPageArray)

		JsonCatPageArray := texajson.SlabToJson(CatPageArray)
		fmt.Println("###JsonCatPageArray: ")
		fmt.Println(JsonCatPageArray)
	}
}

func welcomeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method) //get	request	method
	if r.Method == "GET" {
		t, _ := template.ParseFiles("www/welcome.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
	}
}

// upload logic
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("method:", r.Method)
	if r.Method == "GET" {
		crutime := time.Now().Unix()
		h := md5.New()
		io.WriteString(h, strconv.FormatInt(crutime, 10))
		token := fmt.Sprintf("%x", h.Sum(nil))

		t, _ := template.ParseFiles("login.html")
		t.Execute(w, token)
	} else {
		r.ParseMultipartForm(32 << 20)
		file, handler, err := r.FormFile("uploadfile")
		if err != nil {
			fmt.Println(err)
			return
		}
		AIName = r.FormValue("AIName")
		fmt.Println(AIName)
		defer file.Close()

		fmt.Fprint(w, "ACKNOWLEDGEMENT:\nUploaded the file. Header Info:\n")
		fmt.Fprintf(w, "%v", handler.Header)
		fmt.Fprint(w, "\n\nVISIT: /texa for interrogation.")
		f, err := os.OpenFile("./www/js/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Selected file: ", handler.Filename)
		defer f.Close()
		io.Copy(f, file)
		// http.Redirect(w, r, "/texa", 301)
	}
}

func resultHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	fmt.Println("--TEXA SERVER--")
	fmt.Println("STATUS: INITIATED")
	fmt.Println("ADDR: http://127.0.0.1:3030")

	fs := http.FileServer(http.Dir("www/js"))
	http.Handle("/js/", http.StripPrefix("/js/", fs))

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/texa", texaHandler)
	http.HandleFunc("/result", resultHandler)

	http.ListenAndServe(":3030", nil)
}
