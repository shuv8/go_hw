package main

import (
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type TestCase struct {
	ID        string
	Request   *SearchRequest
	Result    *SearchResponse
	ExpLength int
	IsError   bool
	ExpError  string
	Query     string
}

func TestFindUsers(t *testing.T) {
	cases := []TestCase{
		{
			ID:        "All users",
			Request:   &SearchRequest{Limit: 25},
			ExpLength: 25,
			IsError:   false,
		},
		{
			ID:      "Query in name",
			Request: &SearchRequest{Limit: 25, Query: "Palmer"},
			Result: &SearchResponse{
				Users: []User{{
					ID:     31,
					Name:   "Palmer Scott",
					Age:    37,
					About:  "Elit fugiat commodo laborum quis eu consequat. In velit magna sit fugiat non proident ipsum tempor eu. Consectetur exercitation labore eiusmod occaecat adipisicing irure consequat fugiat ullamco aliquip nostrud anim irure enim. Duis do amet cillum eiusmod eu sunt. Minim minim sunt sit sit enim velit sint tempor enim sint aliquip voluptate reprehenderit officia. Voluptate magna sit consequat adipisicing ut eu qui.\n",
					Gender: "male",
				}},
				NextPage: false,
			},
			ExpLength: 1,
			IsError:   false,
		},
		{
			ID:      "Query in about",
			Request: &SearchRequest{Limit: 25, Query: "cupidatat consequat"},
			Result: &SearchResponse{
				Users: []User{{
					ID:     32,
					Name:   "Christy Knapp",
					Age:    40,
					About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
					Gender: "female",
				}},
				NextPage: false,
			},
			ExpLength: 1,
			IsError:   false,
		},
		{
			ID:        "Limit",
			Request:   &SearchRequest{Limit: 5},
			ExpLength: 5,
			IsError:   false,
		},
		{
			ID:        "Limit > 25",
			Request:   &SearchRequest{Limit: 50},
			ExpLength: 25,
			IsError:   false,
		},
		{
			ID:      "Offset",
			Request: &SearchRequest{Offset: 5, Limit: 1},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     5,
						Name:   "Beulah Stark",
						Age:    30,
						About:  "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			ExpLength: 1,
			IsError:   false,
		},
		{
			ID:      "Next Page",
			Request: &SearchRequest{Limit: 1},
			Result: &SearchResponse{
				Users: []User{{
					ID:     0,
					Name:   "Boyd Wolf",
					Age:    22,
					About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
					Gender: "male",
				}},
				NextPage: true,
			},
			ExpLength: 1,
			IsError:   false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: "test_token_test",
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		assert.Equal(t, item.ExpLength, len(result.Users), item.ID)
		if item.Result != nil {
			assert.Equal(t, item.Result.Users, result.Users, item.ID)
		}
	}
}

func TestOrder(t *testing.T) {
	cases := []TestCase{
		{
			ID:      "Order by Asc Name",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByAsc},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     15,
						Name:   "Allison Valdez",
						Age:    21,
						About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
						Gender: "male",
					}, {
						ID:     16,
						Name:   "Annie Osborn",
						Age:    35,
						About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
						Gender: "female",
					}, {
						ID:     19,
						Name:   "Bell Bauer",
						Age:    26,
						About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
		{
			ID:      "Order by Asc ID",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByAsc, OrderField: "id"},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     0,
						Name:   "Boyd Wolf",
						Age:    22,
						About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
						Gender: "male",
					}, {
						ID:     1,
						Name:   "Hilda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					}, {
						ID:     2,
						Name:   "Brooks Aguilar",
						Age:    25,
						About:  "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
		{
			ID:      "Order by Asc Age",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByAsc, OrderField: "age"},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     1,
						Name:   "Hilda Mayer",
						Age:    21,
						About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
						Gender: "female",
					},
					{
						ID:     15,
						Name:   "Allison Valdez",
						Age:    21,
						About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
						Gender: "male",
					},
					{
						ID:     23,
						Name:   "Gates Spencer",
						Age:    21,
						About:  "Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
		{
			ID:      "Order by Desc Name",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByDesc},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     13,
						Name:   "Whitley Davidson",
						Age:    40,
						About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
						Gender: "male",
					},
					{
						ID:     33,
						Name:   "Twila Snow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
					{
						ID:     18,
						Name:   "Terrell Hall",
						Age:    27,
						About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
		{
			ID:      "Order by Desc ID",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByDesc, OrderField: "id"},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     34,
						Name:   "Kane Sharp",
						Age:    34,
						About:  "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n",
						Gender: "male",
					},
					{
						ID:     33,
						Name:   "Twila Snow",
						Age:    36,
						About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
						Gender: "female",
					},
					{
						ID:     32,
						Name:   "Christy Knapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
		{
			ID:      "Order by Desc Age",
			Request: &SearchRequest{Limit: 3, OrderBy: OrderByDesc, OrderField: "age"},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     13,
						Name:   "Whitley Davidson",
						Age:    40,
						About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
						Gender: "male",
					},
					{
						ID:     32,
						Name:   "Christy Knapp",
						Age:    40,
						About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
						Gender: "female",
					},
					{
						ID:     6,
						Name:   "Jennings Mays",
						Age:    39,
						About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			ExpLength: 3,
			IsError:   false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for caseNum, item := range cases {
		c := &SearchClient{
			AccessToken: "test_token_test",
			URL:         ts.URL,
		}
		result, err := c.FindUsers(*item.Request)

		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		assert.Equal(t, item.ExpLength, len(result.Users), item.ID)
		if item.Result != nil {
			assert.Equal(t, item.Result.Users, result.Users, item.ID)
		}
	}
}

func TestNegativeAuth(t *testing.T) {
	cases := []TestCase{
		{
			ID:      "Unauthorized",
			Request: &SearchRequest{},
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for _, item := range cases {
		c := &SearchClient{
			AccessToken: "wrong_token",
			URL:         ts.URL,
		}
		_, err := c.FindUsers(*item.Request)
		assert.Equal(t, "bad AccessToken", err.Error())
	}
}

func TestDecodingError(t *testing.T) {
	cases := []TestCase{
		{
			ID:       "Decode error",
			Request:  &SearchRequest{},
			ExpError: "SearchServer fatal error",
			IsError:  true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	XMLUserPointer = nil
	_, err := c.FindUsers(*cases[0].Request)
	assert.Equal(t, cases[0].ExpError, err.Error())
}

func TestClientErrors(t *testing.T) {
	cases := []TestCase{
		{
			ID:       "Limit < 0",
			Request:  &SearchRequest{Limit: -1},
			ExpError: "limit must be > 0",
			IsError:  true,
		},
		{
			ID:       "Offset < 0",
			Request:  &SearchRequest{Offset: -1},
			ExpError: "offset must be > 0",
			IsError:  true,
		},
		{
			ID:       "nil URL",
			Request:  &SearchRequest{},
			ExpError: "offset must be > 0",
			IsError:  true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for _, item := range cases {
		if item.ID == "nil URL" {
			c := &SearchClient{
				AccessToken: "test_token",
			}
			_, err := c.FindUsers(*item.Request)
			assert.Contains(t, err.Error(), "unknown error")
		} else {
			c := &SearchClient{
				AccessToken: "test_token",
				URL:         ts.URL,
			}
			_, err := c.FindUsers(*item.Request)
			assert.Equal(t, item.ExpError, err.Error())
		}
	}
}

func TestServerErrors(t *testing.T) {
	cases := []TestCase{
		{
			ID:       "Wrong XML path",
			Request:  &SearchRequest{},
			ExpError: "SearchServer fatal error",
			IsError:  true,
		},
	}

	XMLFilePath = "wrong_path"

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	_, err := c.FindUsers(*cases[0].Request)
	assert.Equal(t, cases[0].ExpError, err.Error())
}

func TestWrongQueryClient(t *testing.T) {
	cases := []TestCase{
		{
			ID:       "Wrong order_field",
			Request:  &SearchRequest{OrderField: "wrong"},
			ExpError: "OrderFeld",
			IsError:  true,
		},
		{
			ID:       "Wrong order_by",
			Request:  &SearchRequest{OrderBy: 123},
			ExpError: "unknown bad request error",
			IsError:  true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	defer ts.Close()

	for _, item := range cases {
		c := &SearchClient{
			AccessToken: "test_token",
			URL:         ts.URL,
		}
		_, err := c.FindUsers(*item.Request)
		assert.Contains(t, err.Error(), item.ExpError)
	}
}

func TestWrongQueryServer(t *testing.T) {
	cases := []TestCase{
		{
			ID:       "Wrong limit",
			Query:    "limit=wrong",
			ExpError: "{\"Error\":\"Wrong limit, should be int\"}",
		},
		{
			ID:       "Wrong offset",
			Query:    "offset=wrong",
			ExpError: "{\"Error\":\"Wrong offset, should be int\"}",
		},
	}
	for caseNum, item := range cases {
		url := "https://example.com/test?" + item.Query
		req := httptest.NewRequest("GET", url, nil)
		req.Header.Add("AccessToken", "test_token")
		w := httptest.NewRecorder()

		SearchServer(w, req)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("[%d] failed to read body: %v", caseNum, err)
		}

		bodyStr := string(body)
		assert.Equal(t, item.ExpError, bodyStr)
	}
}

func TestNegativeSerializeData(t *testing.T) {
	result := SerializeData(make(chan int))
	assert.Equal(t, result, []byte(nil))
}

func wrongHandler(w http.ResponseWriter, r *http.Request) {
	data := SerializeData("test")
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	SendResponse(w, &data, http.StatusNoContent)
}

func TestNegativeSendResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(wrongHandler))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})
	assert.Contains(t, err.Error(), "EOF")
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(10 * time.Second)
	if _, err := w.Write([]byte("hello")); err != nil {
		log.Println(err)
	}
}

func TestNegativeClientTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(slowHandler))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})
	assert.Contains(t, err.Error(), "timeout for")
}

func badJSONHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Error in responding", http.StatusOK)
}

func TestNegativeClientUnmarshal(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(badJSONHandler))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})
	assert.Contains(t, err.Error(), "cant unpack result json")
}

func badJSONErrorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Error in responding", http.StatusBadRequest)
}

func TestNegativeClientUnmarshalError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(badJSONErrorHandler))
	defer ts.Close()

	c := &SearchClient{
		AccessToken: "test_token",
		URL:         ts.URL,
	}
	_, err := c.FindUsers(SearchRequest{})
	assert.Contains(t, err.Error(), "cant unpack error json")
}
