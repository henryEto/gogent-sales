package main

import (
	"context"
	"copo-ai-agent/internal/database"
	"encoding/json"
	"log"
)

func getCodesList(queries *database.Queries, args map[string]any) string {
	productos, err := queries.GetAllProductCodes(context.Background())
	if err != nil {
		log.Printf("failed to get products list: %v", err)
		return "ocurrió un error al obtener la lista de códigos"
	}

	jsonData, err := json.Marshal(productos)
	if err != nil {
		log.Printf("failed to marshal results: %v", err)
		return "ocurrió un error al obtener la lista de códigos"
	}

	return string(jsonData)
}

func getProductInfoBySearchTerm(queries *database.Queries, args map[string]any) string {
	arg := args["searchTerm"]
	searchTerm, ok := arg.(string)
	if !ok {
		log.Println("failed to extract argument for term based search...")
		return "ocurrió un problema al obtener la lista de códigos por búsqueda"
	}
	codigos, err := queries.GetProductCodesBySearchTerm(
		context.Background(),
		database.GetProductCodesBySearchTermParams{SearchTerm: searchTerm},
	)
	if err != nil {
		log.Printf("failed to get products list by search term: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por búsqueda"
	}

	var codigosString []string
	for _, c := range codigos {
		codigosString = append(codigosString, c.Codigo)
	}
	mapCodigos := map[string]any{
		"productCodes": codigosString,
	}

	info := getProductsInfo(queries, mapCodigos)

	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Printf("failed to marshal products list by search: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por búsqueda"
	}

	return string(jsonData)
}

func getProductInfoByBrand(queries *database.Queries, args map[string]any) string {
	brandArg := args["brand"]
	brand, ok := brandArg.(string)
	if !ok {
		log.Println("failed to extract argument for brand based search...")
		return "ocurrió un problema al obtener la lista de códigos por brand"
	}
	codigos, err := queries.GetProductCodesByBrand(
		context.Background(),
		brand,
	)
	if err != nil {
		log.Printf("failed to get products list by brand: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por marca"
	}

	var codigosString []string
	for _, c := range codigos {
		codigosString = append(codigosString, c.Codigo)
	}
	mapCodigos := map[string]any{
		"productCodes": codigosString,
	}

	info := getProductsInfo(queries, mapCodigos)

	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Printf("failed to marshal products list by brand: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por marca"
	}

	return string(jsonData)
}

func getProductInfoByCategories(queries *database.Queries, args map[string]any) string {
	lineaArg := args["linea"]
	sublineaArg := args["sublinea"]
	linea, ok := lineaArg.(string)
	if !ok {
		log.Println("failed to extract argument for linea based search...")
		return "ocurrió un problema al obtener la lista de códigos por linea"
	}
	sublinea, ok := sublineaArg.(string)
	if !ok {
		log.Println("failed to extract argument for sublinea based search...")
		return "ocurrió un problema al obtener la lista de códigos por sbulinea"
	}

	codigos, err := queries.GetProductCodesByCategory(
		context.Background(),
		database.GetProductCodesByCategoryParams{
			Linea:    linea,
			Sublinea: sublinea,
		},
	)
	if err != nil {
		log.Printf("failed to get products list by category: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por linea"
	}

	var codigosString []string
	for _, c := range codigos {
		codigosString = append(codigosString, c.Codigo)
	}
	mapCodigos := map[string]any{
		"productCodes": codigosString,
	}

	info := getProductsInfo(queries, mapCodigos)

	jsonData, err := json.Marshal(info)
	if err != nil {
		log.Printf("failed to marshal products list by category: %v", err)
		return "ocurrió un problema al obtener la lista de códigos por linea"
	}

	return string(jsonData)
}

func getProductsInfo(queries *database.Queries, args map[string]any) string {
	var productCodes []string
	if argCodes, ok := args["productCodes"]; ok {
		if codeSlice, ok := argCodes.([]string); ok {
			productCodes = codeSlice
		} else if codesSlice, ok := argCodes.([]any); ok {
			log.Println("argument is of type []any")
			for _, v := range codesSlice {
				if str, ok := v.(string); ok {
					log.Println(str)
					productCodes = append(productCodes, str)
				}
			}
		}
	} else {
		log.Println("failed to extract argument for codigos...")
		return "ocurrió un problema al obtener información de los códigos"
	}

	infoProductos, err := queries.GetProductsInfoByCode(context.Background(), productCodes)
	if err != nil {
		log.Printf("failed to get products info: %v", err)
		return "ocurrió un error al obtener la información de los productos"
	}

	jsonData, err := json.Marshal(infoProductos)
	if err != nil {
		log.Printf("failed to marshal results: %v", err)
		return "ocurrió un error al obtener la información de los productos"
	}

	return string(jsonData)
}
