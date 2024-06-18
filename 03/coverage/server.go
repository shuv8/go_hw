package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type XMLUser struct {
	ID     int    `xml:"id"`
	FName  string `xml:"first_name"`
	SName  string `xml:"last_name"`
	Age    int    `xml:"age"`
	About  string `xml:"about"`
	Gender string `xml:"gender"`
}

var (
	XMLFilePath      = "dataset.xml"
	XMLUserPointer   = new(XMLUser)
	validOrderField  = regexp.MustCompile(`^(id|age|name|)$`)
	validAccessToken = regexp.MustCompile(`^test_token.*$`)
)

func GetUsersFromXML(query string) ([]User, error) {
	f, err := os.Open(XMLFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)

	users := make([]User, 0)
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch tp := tok.(type) {
		case xml.StartElement:
			if tp.Name.Local == "row" {
				decodeErr := decoder.DecodeElement(XMLUserPointer, &tp)
				if decodeErr != nil {
					return nil, errors.New("failed decode")
				}
				user := User{
					ID:     XMLUserPointer.ID,
					Name:   strings.Join([]string{XMLUserPointer.FName, XMLUserPointer.SName}, " "),
					Age:    XMLUserPointer.Age,
					About:  XMLUserPointer.About,
					Gender: XMLUserPointer.Gender,
				}
				if query == "" || strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
					users = append(users, user)
				}
			}
		default:
			break
		}
	}

	return users, nil
}

func parseQuery(requestQuery url.Values) (SearchRequest, SearchErrorResponse) {
	orderField := requestQuery.Get("order_field")
	if !validOrderField.MatchString(orderField) {
		return SearchRequest{}, SearchErrorResponse{Error: ErrorBadOrderField}
	}
	orderBy := requestQuery.Get("order_by")
	orderByInt := OrderByAsIs
	if orderBy != "" {
		var err error
		orderByInt, err = strconv.Atoi(orderBy)
		if err != nil || (orderByInt != OrderByAsc && orderByInt != OrderByDesc && orderByInt != OrderByAsIs) {
			return SearchRequest{}, SearchErrorResponse{Error: "Wrong order_by, should be 1, 0 or -1"}
		}
	}
	limit := requestQuery.Get("limit")
	limitInt := 0
	if limit != "" {
		var err error
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			return SearchRequest{}, SearchErrorResponse{Error: "Wrong limit, should be int"}
		}
	}
	offset := requestQuery.Get("offset")
	offsetInt := 0
	if offset != "" {
		var err error
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			return SearchRequest{}, SearchErrorResponse{Error: "Wrong offset, should be int"}
		}
	}
	query := requestQuery.Get("query")
	return SearchRequest{
		Limit:      limitInt,
		Offset:     offsetInt,
		Query:      query,
		OrderField: orderField,
		OrderBy:    orderByInt,
	}, SearchErrorResponse{}
}

func SerializeData(data any) []byte {
	result, marshalErr := json.Marshal(data)
	if marshalErr != nil {
		return nil
	}
	return result
}

func SendResponse(w http.ResponseWriter, data *[]byte, statusCode int) {
	w.WriteHeader(statusCode)
	_, responseErr := w.Write(*data)
	if responseErr != nil {
		http.Error(w, "Error in responding", http.StatusInternalServerError)
	}
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	accToken := r.Header.Get("AccessToken")
	if !validAccessToken.MatchString(accToken) {
		http.Error(w, "Unauthorized!", http.StatusUnauthorized)
		return
	}
	requestQuery := r.URL.Query()
	request, queryError := parseQuery(requestQuery)
	if queryError.Error != "" {
		result := SerializeData(queryError)
		SendResponse(w, &result, http.StatusBadRequest)
		return
	}

	users, err := GetUsersFromXML(request.Query)
	if err != nil {
		http.Error(w, "Error in getting users!", http.StatusInternalServerError)
		return
	}

	sort.SliceStable(users, func(i, j int) bool {
		switch request.OrderBy {
		case OrderByAsc:
			switch request.OrderField {
			case "id":
				return users[i].ID < users[j].ID
			case "age":
				return users[i].Age < users[j].Age
			default:
				return users[i].Name < users[j].Name
			}
		case OrderByDesc:
			switch request.OrderField {
			case "id":
				return users[i].ID > users[j].ID
			case "age":
				return users[i].Age > users[j].Age
			default:
				return users[i].Name > users[j].Name
			}
		}
		return false
	})

	if request.Offset != 0 {
		users = users[request.Offset:]
	}
	if request.Limit != 0 && request.Limit < len(users) {
		users = users[:request.Limit]
	}

	w.Header().Set("Content-Type", "application/json")
	result := SerializeData(users)
	SendResponse(w, &result, http.StatusOK)
}
