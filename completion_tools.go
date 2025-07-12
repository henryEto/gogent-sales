package main

import (
	"copo-ai-agent/internal/database"

	"google.golang.org/genai"
)

type CompletionTools struct {
	Tools []FunctionTool
}

type FunctionTool struct {
	Name        string
	Declaration *genai.FunctionDeclaration
	Function    func(*database.Queries, map[string]any) string
}

func getCompletionTools() CompletionTools {
	return CompletionTools{
		Tools: []FunctionTool{
			{
				Name:     "obtenerListaProductos",
				Function: getCodesList,
				Declaration: &genai.FunctionDeclaration{
					Name:        "obtenerListaProductos",
					Description: "Devuelve un JSON con la lista completa de los productos disponbles. Incluye codigo y descripción del producto",
					Parameters:  &genai.Schema{Type: genai.TypeObject},
					Response:    &genai.Schema{Type: genai.TypeString},
				},
			},
			{
				Name:     "obtenerInformacionPorBusqueda",
				Function: getProductInfoBySearchTerm,
				Declaration: &genai.FunctionDeclaration{
					Name: "obtenerInformacionPorBusqueda",
					Description: "Hace una búsqueda de productos basado en un término de búsqueda " +
						"y devuelve un JSON con la información detallada de los productos. El término de búsqueda debe ser una sola palabra en singular.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"searchTerm": {
								Type:        genai.TypeString,
								Description: "El término para realizar la busqueda. Debe ser una sola palabra en singular",
							},
						},
						Required: []string{"searchTerm"},
					},
					Response: &genai.Schema{Type: genai.TypeString},
				},
			},
			{
				Name:     "obtenerInformacionPorMarca",
				Function: getProductInfoByBrand,
				Declaration: &genai.FunctionDeclaration{
					Name: "obtenerInformacionPorMarca",
					Description: "Hace una búsqueda de productos por marca " +
						"y devuelve un JSON con la información detallada de los productos: " +
						"descripción, línea, sublínea, marca, existencia, popularidad, pesos promedio, " +
						"piezas por caja, y precios.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"brand": {
								Type:        genai.TypeString,
								Description: "La marca para realizar la busqueda",
							},
						},
						Required: []string{"brand"},
					},
					Response: &genai.Schema{Type: genai.TypeString},
				},
			},
			{
				Name:     "obtenerInformacionPorLineaSublinea",
				Function: getProductInfoByCategories,
				Declaration: &genai.FunctionDeclaration{
					Name: "obtenerInformacionPorLineaSublinea",
					Description: "Hace una búsqueda de productos por línea y sublinea " +
						"y devuelve un JSON con la información detallada de los productos: " +
						"descripción, línea, sublínea, marca, existencia, popularidad, pesos promedio, " +
						"piezas por caja, y precios.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"linea": {
								Type:        genai.TypeString,
								Description: "La linea para realizar la busqueda, si se quieren buscar todas la lineas debe ser un texto vacio ''",
							},
							"sublinea": {
								Type:        genai.TypeString,
								Description: "La linea para realizar la busqueda, si se quieren buscar todas la sublineas debe ser un texto vacio ''",
							},
						},
						Required: []string{"linea", "sublinea"},
					},
					Response: &genai.Schema{Type: genai.TypeString},
				},
			},
			{
				Name:     "obtenerInformacionPorCodigo",
				Function: getProductsInfo,
				Declaration: &genai.FunctionDeclaration{
					Name: "obtenerInformacionPorCodigo",
					Description: "Devuelve un JSON con info. detallada de productos por código: " +
						"descripción, línea, sublínea, marca, existencia, popularidad, pesos promedio, " +
						"piezas por caja, y precios escalonados.",
					Parameters: &genai.Schema{
						Type: genai.TypeObject,
						Properties: map[string]*genai.Schema{
							"productCodes": {
								Type:        genai.TypeArray,
								Description: "Lista de códigos de productos (strings).",
								Items:       &genai.Schema{Type: genai.TypeString},
							},
						},
						Required: []string{"productCodes"},
					},
					Response: &genai.Schema{Type: genai.TypeString},
				},
			},
		},
	}
}

func (ct *CompletionTools) getDeclarationsList() []*genai.FunctionDeclaration {
	var listFD []*genai.FunctionDeclaration
	for _, tool := range ct.Tools {
		listFD = append(listFD, tool.Declaration)
	}
	return listFD
}

func (ct *CompletionTools) getToolByName(name string) FunctionTool {
	for _, tool := range ct.Tools {
		if tool.Name == name {
			return tool
		}
	}
	return FunctionTool{} // Return an empty FunctionTool if not found
}
