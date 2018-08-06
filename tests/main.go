package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"code.videolan.org/videolan/CrashDragon/database"
)

type apiProdResp struct {
	Error string
	Item  database.Product
}

type apiProdListResp struct {
	Error string
	Items []database.Product
}

func testProduct() {
	client := &http.Client{}

	var Product database.Product
	Product.Name = "Testproduct 1"
	Product.Slug = "tp1"
	j, err := json.Marshal(Product)
	if err != nil {
		log.Panic(err)
	}

	// Test POST
	res, err := http.Post("http://admin:12345@127.0.0.1:8080/api/v1/products", "application/json", strings.NewReader(string(j)))
	if err != nil {
		log.Panic(err)
	}
	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Panic(err)
	}
	var Product2 apiProdResp
	err = json.Unmarshal(s, &Product2)
	if err != nil {
		log.Panic(err)
	}
	if Product2.Item.Name == Product.Name && Product2.Item.Slug == Product.Slug {
		log.Println("Product POSTed with ID", Product2.Item.ID.String())
	} else {
		log.Println("Should:", Product)
		log.Println("Is:", Product2.Item)
		log.Fatalln("Product could not be POSTed, error:", Product2.Error)
	}

	// Test LIST
	res2, err := http.Get("http://admin:12345@127.0.0.1:8080/api/v1/products")
	if err != nil {
		log.Panic(err)
	}
	s2, err := ioutil.ReadAll(res2.Body)
	if err != nil {
		log.Panic(err)
	}
	var Product3 apiProdListResp
	err = json.Unmarshal(s2, &Product3)
	if err != nil {
		log.Panic(err)
	}
	found := false
	for _, prod := range Product3.Items {
		if prod.Name == Product.Name && prod.Slug == Product.Slug {
			found = true
			log.Println("Product found with ID", prod.ID.String())
		}
	}
	if found == false {
		log.Println("Should:", Product)
		log.Println("Is:", Product3.Items)
		log.Fatalln("Product could not be found, error:", Product3.Error)
	}

	// Test GET
	res3, err := http.Get("http://admin:12345@127.0.0.1:8080/api/v1/products/" + Product2.Item.ID.String())
	if err != nil {
		log.Panic(err)
	}
	s3, err := ioutil.ReadAll(res3.Body)
	if err != nil {
		log.Panic(err)
	}
	var Product4 apiProdResp
	err = json.Unmarshal(s3, &Product4)
	if err != nil {
		log.Panic(err)
	}
	if Product4.Item.Name == Product.Name && Product4.Item.Slug == Product.Slug {
		log.Println("Product GOT with ID", Product4.Item.ID.String())
	} else {
		log.Println("Should:", Product)
		log.Println("Is:", Product4.Item)
		log.Fatalln("Product could not be GOT, error:", Product4.Error)
	}

	// Test PUT
	Product.Name = "Testproduct 2"
	Product.Slug = "tp2"
	j, err = json.Marshal(Product)
	if err != nil {
		log.Panic(err)
	}

	req, err := http.NewRequest("PUT", "http://admin:12345@127.0.0.1:8080/api/v1/products/"+Product2.Item.ID.String(), strings.NewReader(string(j)))
	if err != nil {
		log.Panic(err)
	}
	res4, err := client.Do(req)
	if err != nil {
		log.Panic(err)
	}
	s4, err := ioutil.ReadAll(res4.Body)
	if err != nil {
		log.Panic(err)
	}
	var Product5 apiProdResp
	err = json.Unmarshal(s4, &Product5)
	if err != nil {
		log.Panic(err)
	}
	if Product5.Item.Name == Product.Name && Product5.Item.Slug == Product.Slug {
		log.Println("Product PUTed with ID", Product5.Item.ID.String())
	} else {
		log.Println("Should:", Product)
		log.Println("Is:", Product5.Item)
		log.Fatalln("Product could not be PUTed, error:", Product5.Error)
	}

	// Test DELETE
	req2, err := http.NewRequest("DELETE", "http://admin:12345@127.0.0.1:8080/api/v1/products/"+Product2.Item.ID.String(), nil)
	if err != nil {
		log.Panic(err)
	}
	res5, err := client.Do(req2)
	if err != nil {
		log.Panic(err)
	}
	s5, err := ioutil.ReadAll(res5.Body)
	if err != nil {
		log.Panic(err)
	}
	var Product6 apiProdResp
	err = json.Unmarshal(s5, &Product6)
	if err != nil {
		log.Panic(err)
	}
	if Product6.Error == "" {
		log.Println("Product DELETEd with ID", Product2.Item.ID.String())
	} else {
		log.Fatalln("Product could not be DELETEd, error:", Product6.Error)
	}
}

func runTests() {
	testProduct()
}

func main() {
	log.Println("Running CrashDragon tests...")
	runTests()
	log.Println("All tests finished successfully")
}
