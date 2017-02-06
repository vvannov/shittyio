package train

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"
)

type Events []string

func (events *Events) Handler(event string, status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		*events = append(*events, event)
		w.WriteHeader(status)
		fmt.Fprint(w, body)
	}
}

func (events Events) CompareWithEthalon(ethalon ...string) bool {
	return reflect.DeepEqual(events, ethalon)
}

func TestTrainBasic(t *testing.T) {
	events := []string{}
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/page1", func(w http.ResponseWriter, r *http.Request) {
			t.Log("cliet reached /page1")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, "OK")
			events = append(events, "/page1")
		})

		trn := New(mux)

		{
			vagon1 := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				t.Log("vagon 1 launched")
				defer t.Log("vagon 1 stopped(defer)")
				events = append(events, "vagon 1")
				next(w, r)
				t.Log("vagon 1 stopped")
			}
			trn.AddVagon(vagon1)
		}

		{
			vagon2 := func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				t.Log("vagon 2 launched")
				defer t.Log("vagon 2 stopped(defer)")
				events = append(events, "vagon 2")
				next(w, r)
				t.Log("vagon 2 stopped")
			}
			trn.AddVagon(vagon2)
		}

		if err := http.ListenAndServe(":9999", trn.Handler()); err != nil {
			t.Error(err)
			t.FailNow()
		}
	}()
	response, err := http.Get("http://localhost:9999/page1")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if response.StatusCode != http.StatusOK {
		t.Log(response.Status)
		t.FailNow()
	}
	defer response.Body.Close()
	//	bufer := make([]byte, 100)
	//	_, err = response.Body.Read(bufer)

	//	body, err := ioutil.ReadAll(response.Body)
	body := new(bytes.Buffer)
	_, err = body.ReadFrom(response.Body)
	if err != nil && err != io.EOF {
		t.Log(err)
		t.FailNow()
	}
	bodyVampire := body.String()
	if bodyVampire != "OK" {
		t.Log(fmt.Sprintf("%#v", bodyVampire))
		t.FailNow()
	}
	if !reflect.DeepEqual(events, []string{"vagon 1", "vagon 2", "/page1"}) {
		t.Log("events isn't the same like etalon", events)
		t.FailNow()
	}
}
