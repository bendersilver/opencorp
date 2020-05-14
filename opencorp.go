package opencorp

import (
	"compress/bzip2"
	"encoding/gob"
	"encoding/xml"
	"log"
	"net/http"
	"os"
)

const dictURL = "http://opencorpora.org/files/export/dict/dict.opcorpora.xml.bz2"

// Data -
type Data struct {
	Grammeme map[string]*Grammeme
	Lemms    []*LemmaF
}

// LemmaForms -
type LemmaForms struct {
	Lemma  string
	Gramms []*Grammeme
}

// LemmaF -
type LemmaF struct {
	ID     int64
	Lemma  string
	Gramms []*Grammeme
	Forms  []*LemmaForms
}

// Grammeme represents opencorpora grammeme
type Grammeme struct {
	Name        string `xml:"name"`
	Alias       string `xml:"alias"`
	Description string `xml:"description"`
	Parent      string `xml:"parent,attr"`
}

// Lemma represents lemma
type Lemma struct {
	ID    int64  `xml:"id,attr"`
	Main  Form   `xml:"l"`
	Forms []Form `xml:"f"`
}

// Form represents lemma form
type Form struct {
	Value         string         `xml:"t,attr"`
	GrammemeNames []GrammemeName `xml:"g"`
}

// GrammemeName represents name of grammeme in lemmas
type GrammemeName struct {
	Value string `xml:"v,attr"`
}

// Update -
func Update() error {
	d := new(Data)
	d.Grammeme = make(map[string]*Grammeme)

	resp, err := http.Get(dictURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(bzip2.NewReader(resp.Body))

	for {
		// exit loop if nothing to decode
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "grammeme" {
				var g *Grammeme
				if err = decoder.DecodeElement(&g, &se); err != nil {
					return err
				}
				d.Grammeme[g.Name] = g
			}
			if se.Name.Local == "lemma" {
				var l Lemma
				if err = decoder.DecodeElement(&l, &se); err != nil {
					return err
				}
				lf := &LemmaF{
					ID:    l.ID,
					Lemma: l.Main.Value,
				}
				for _, v := range l.Main.GrammemeNames {
					lf.Gramms = append(lf.Gramms, d.Grammeme[v.Value])
				}
				for _, v := range l.Forms {
					f := &LemmaForms{
						Lemma: v.Value,
					}
					for _, fr := range v.GrammemeNames {
						f.Gramms = append(f.Gramms, d.Grammeme[fr.Value])
					}
					lf.Forms = append(lf.Forms, f)
				}
				d.Lemms = append(d.Lemms, lf)
				// log.Print(lf)

				// time.Sleep(time.Second / 10)
			}
		}
	}
	gb, err := os.Create("data.gob")
	if err != nil {
		return err
	}
	defer gb.Close()
	g := gob.NewEncoder(gb)
	g.Encode(d)
	log.Print("ok")
	return nil
}

func init() {
	Update()
}
