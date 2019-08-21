package gql

import (
	"github.com/dusk-network/dusk-blockchain/pkg/core/database"
	"github.com/dusk-network/dusk-blockchain/pkg/gql/query"
	"io/ioutil"
	"net/http"

	"github.com/go-chi/render"
	"github.com/graphql-go/graphql"
)

// handleQuery to process graphQL query
func handleQuery(schema *graphql.Schema, w http.ResponseWriter, r http.Request, db database.DB) {

	if r.Body == nil {
		http.Error(w, "Must provide graphql query in request body", 400)
		return
	}

	decBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Execute graphql query
	result := query.Execute(string(decBody), *schema, db)
	render.JSON(w, &r, result)
}
