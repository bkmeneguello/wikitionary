package main

import (
	"flag"
	"os"
	"bufio"
	"encoding/xml"
	"compress/bzip2"
	"io"
	"log"
	"fmt"
	"strings"
	"regexp"
)

type Page struct {
	Title, Text string
}

func ExtractPage(file_name string) (out chan *Page) {
	out = make(chan *Page)
	go func() {
		file, err := os.Open(file_name)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		bz2file := bzip2.NewReader(bufio.NewReaderSize(file, 16384))

		decoder := xml.NewDecoder(bz2file)
		page := false
		title := false
		title_cd := ""
		ns := false
		ns_cd := ""
		text := false
		text_cd := ""
		for {
			t, err := decoder.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			switch t1 := t.(type) {
			case xml.StartElement:
				switch t.(xml.StartElement).Name.Local {
				case "page":
					page = true
				case "title":
					title = true
				case "ns":
					ns = true
				case "text":
					text = true
				}
			case xml.EndElement:
				switch t.(xml.EndElement).Name.Local {
				case "page":
					page = false
					if ns_cd == "0" {
						out <- &Page{Title: title_cd, Text: text_cd}
					}
				case "title":
					title = false
				case "ns":
					ns = false
				case "text":
					text = false
				}
			case xml.CharData:
				if page {
					if title {
						title_cd = string(t.(xml.CharData))
					} else if ns {
						ns_cd = string(t.(xml.CharData))
					} else if text {
						text_cd = string(t.(xml.CharData))
					}
				}
			default:
				log.Fatalf("%T\n", t1)
			}
		}
		close(out)
	}()
	return
}

type Word struct {
	string
	Type string
}

func ParseTest(in <-chan *Page) (out chan *Word) {
	out = make(chan *Word)
	go func() {
		for page := range in {
			text := page.Text
			scanner := bufio.NewScanner(strings.NewReader(text))
			scanner.Split(bufio.ScanLines)
			lang_pt := false
			class := ""
			lang_regex := regexp.MustCompile("=\\{\\{-([a-z]{2})-\\}\\}=")
			class_regex := regexp.MustCompile("==([^=]+)==")
			cat_regex := regexp.MustCompile("\\[\\[Categoria:([^=]+)\\(Português\\)\\]\\]")
			found := false
			classes := []string{}
			puts := func(class string) {
				contains := false
				for _, c := range classes {
					if class == c {
						contains = true
						break
					}
				}
				if ! contains {
					classes = append(classes, class)
					w := &Word{string: page.Title, Type: class}
					found = true
					out <- w
				}
			}
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "={{") {
					if lang_sub := lang_regex.FindStringSubmatch(line); len(lang_sub) > 0 {
						lang_pt = lang_sub[1] == "pt"
					}
				}
				if lang_pt {
					if strings.HasPrefix(line, "==") {
						if class_sub := class_regex.FindStringSubmatch(line); len(class_sub) > 0 {
							class = strings.ToLower(strings.Trim(class_sub[1], " "))
						} else {
							continue
						}
					} else if (strings.HasPrefix(line, "[[Categoria:")) {
						if cat_sub := cat_regex.FindStringSubmatch(line); len(cat_sub) > 0 {
							class = strings.ToLower(strings.Trim(cat_sub[1], " "))
						} else {
							continue
						}
					} else {
						continue
					}

					switch class {
					case "substantivo1": //???
						fallthrough
					case "substantivo2": //???
						fallthrough
					case "substantivo<sup>1</sup>": //???
						fallthrough
					case "substantivo<sup>2</sup>": //???
						fallthrough
					case "substantivo, ''feminino''": //??? !!!
						puts("substantivo")
					case "adjetivo<sup>1</sup>": //???
						fallthrough
					case "adjetivo<sup>2</sup>": //???
						fallthrough
					case "adjetiivo": //??? !!!
						puts("adjetivo")
					case "forma de sufixo1": //???
						fallthrough
					case "forma de sufixo<sup>1</sup>": //???
						fallthrough
					case "forma de sufixo2": //???
						fallthrough
					case "forma de sufixo<sup>2</sup>": //???
						puts("forma de sufixo")
					case "verbo <sup>1</sup>": //???
						fallthrough
					case "verbo<sup>1</sup>": //???
						fallthrough
					case "verbo <sup>(1)</sup>": //???
						fallthrough
					case "verbo <sup>2</sup>": //???
						fallthrough
					case "verbo<sup>2</sup>": //???
						fallthrough
					case "verbo <sup>(2)</sup>": //???
						fallthrough
					case "verbo¹": //???
						fallthrough
					case "verbo²": //???
						puts("verbo")
					case "sigla<sup>1</sup>": //???
						fallthrough
					case "sigla<sup>2</sup>": //???
						fallthrough
					case "sigla<sup>3</sup>": //???
						puts("sigla")
					case "adjetivo/substantivo": //???
						puts("adjetivo")
						puts("substantivo")
					case "locução substantiva1": //???
						fallthrough
					case "locução substantiva2": //???
						puts("locução substantiva")
					case "sufixo1": //???
						fallthrough
					case "sufixo2": //???
						puts("sufixo")
					case "forma de sigla<sup>1</sup>": //???
						fallthrough
					case "forma de sigla<sup>2</sup>": //???
						puts("forma de sigla")
					case "advérbio<sup>1</sup>": //???
						fallthrough
					case "advérbio<sup>2</sup>": //???
						puts("advérbio")
					case "substantivo próprio": //???
						fallthrough
					case "substantivo <small>próprio</small>": //???
						puts("substantivo próprio")
					case "forma de substantivo": //???
						fallthrough
					case "forma de locução substantiva": //???
						fallthrough
					case "locução": //???
						fallthrough
					case "forma de adjetivo": //???
						fallthrough
					case "locução interjetiva": //???
						fallthrough
					case "forma de locução adverbial": //???
						fallthrough
					case "forma de advérbio": //???
						fallthrough
					case "numeral cardinal": //???
						fallthrough
					case "substantivo comum": //???
						fallthrough
					case "numeral ordinal": //???
						fallthrough
					case "abreviação": //???
						fallthrough
					case "antepositivo": //???
						fallthrough
					case "locucao pronominal": //???
						fallthrough
					case "pronome pessoal": //???
						fallthrough
					case "locução pronominal": //???
						fallthrough
					case "forma de pronome": //???
						fallthrough
					case "locução conjuntiva": //???
						fallthrough
					case "topónimo": //???
						fallthrough
					case "locução verbal": //???
						fallthrough
					case "elemento de composição": //???
						fallthrough
					case "forma de sigla": //???
						fallthrough
					case "forma de locução": //???
						fallthrough
					case "onomatopeia": //???
						fallthrough
					case "numeral multiplicativo": //???
						fallthrough
					case "forma de verbo": //???
						fallthrough
					case "infixo": //???
						fallthrough
					case "pospositivo": //???
						fallthrough
					case "forma de sufixo": //???
						fallthrough
					case "interfixo": //???
						fallthrough
					case "terminação": //???
						fallthrough
					case "forma de locução pronominal": //???
						fallthrough
					case "gramática": //???
						//skip
					case "afixo": //???
						fallthrough
					case "frase": //???
						//skip
					case "artigo":
						fallthrough
					case "adjetivo":
						fallthrough
					case "advérbio":
						fallthrough
					case "conjunção":
						fallthrough
					case "interjeição":
						fallthrough
					case "numeral":
						fallthrough
					case "partícula":
						fallthrough
					case "preposição":
						fallthrough
					case "posposição":
						fallthrough
					case "pronome":
						fallthrough
					case "substantivo":
						fallthrough
					case "verbo":
						fallthrough
					case "forma verbal":
						fallthrough
					//------------------
					case "locução substantiva":
						fallthrough
					case "locução adjetiva":
						fallthrough
					case "locução adverbial":
						fallthrough
					case "locução prepositiva":
						fallthrough
					case "expressão":
						fallthrough
					//------------------
					case "abreviatura":
						fallthrough
					case "contração":
						fallthrough
					case "prefixo":
						fallthrough
					case "sufixo":
						fallthrough
					case "sigla":
						fallthrough
					case "símbolo":
						puts(class)
					//------------------
					default:
						//log.Println(page.Title, class)
					}
				}
			}
			if lang_pt && !found {
				//log.Println(page)
			}
		}
		close(out)
	}()
	return
}

func main() {
	flag.Parse()

	page_stream := ExtractPage(flag.Arg(0))
	page2_stream := ParseTest(page_stream)
	count := 0
	for w := range page2_stream {
		count++
		fmt.Printf("\"%s\" [%s]\n", w.string, w.Type)
	}
	fmt.Println(count)
}
