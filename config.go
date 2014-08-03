//  Copyright (c) 2014 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.
package bleve

import (
	"fmt"
	"regexp"
	"time"

	"github.com/couchbaselabs/bleve/analysis"

	"github.com/couchbaselabs/bleve/analysis/datetime_parsers/flexible_go"

	"github.com/couchbaselabs/bleve/analysis/char_filters/regexp_char_filter"

	"github.com/couchbaselabs/bleve/analysis/tokenizers/regexp_tokenizer"
	"github.com/couchbaselabs/bleve/analysis/tokenizers/single_token"
	"github.com/couchbaselabs/bleve/analysis/tokenizers/unicode_word_boundary"

	"github.com/couchbaselabs/bleve/analysis/token_filters/cld2"
	"github.com/couchbaselabs/bleve/analysis/token_filters/length_filter"
	"github.com/couchbaselabs/bleve/analysis/token_filters/lower_case_filter"
	"github.com/couchbaselabs/bleve/analysis/token_filters/stemmer_filter"
	"github.com/couchbaselabs/bleve/analysis/token_filters/stop_words_filter"

	"github.com/couchbaselabs/bleve/search"
)

type AnalysisConfig struct {
	StopTokenMaps   map[string]stop_words_filter.StopWordsMap
	CharFilters     map[string]analysis.CharFilter
	Tokenizers      map[string]analysis.Tokenizer
	TokenFilters    map[string]analysis.TokenFilter
	Analyzers       map[string]*analysis.Analyzer
	DateTimeParsers map[string]analysis.DateTimeParser
}

type HighlightConfig struct {
	Highlighters map[string]search.Highlighter
}

type Configuration struct {
	Analysis              *AnalysisConfig
	DefaultAnalyzer       *string
	Highlight             *HighlightConfig
	DefaultHighlighter    *string
	CreateIfMissing       bool
	DefaultDateTimeFormat *string
}

func (c *Configuration) BuildNewAnalyzer(charFilterNames []string, tokenizerName string, tokenFilterNames []string) (*analysis.Analyzer, error) {
	rv := analysis.Analyzer{}
	if len(charFilterNames) > 0 {
		rv.CharFilters = make([]analysis.CharFilter, len(charFilterNames))
		for i, charFilterName := range charFilterNames {
			charFilter := c.Analysis.CharFilters[charFilterName]
			if charFilter == nil {
				return nil, fmt.Errorf("no character filter named `%s` registered", charFilterName)
			}
			rv.CharFilters[i] = charFilter
		}
	}
	rv.Tokenizer = c.Analysis.Tokenizers[tokenizerName]
	if rv.Tokenizer == nil {
		return nil, fmt.Errorf("no tokenizer named `%s` registered", tokenizerName)
	}
	if len(tokenFilterNames) > 0 {
		rv.TokenFilters = make([]analysis.TokenFilter, len(tokenFilterNames))
		for i, tokenFilterName := range tokenFilterNames {
			tokenFilter := c.Analysis.TokenFilters[tokenFilterName]
			if tokenFilter == nil {
				return nil, fmt.Errorf("no token filter named `%s` registered", tokenFilterName)
			}
			rv.TokenFilters[i] = tokenFilter
		}
	}
	return &rv, nil
}

func (c *Configuration) MustBuildNewAnalyzer(charFilterNames []string, tokenizerName string, tokenFilterNames []string) *analysis.Analyzer {
	analyzer, err := c.BuildNewAnalyzer(charFilterNames, tokenizerName, tokenFilterNames)
	if err != nil {
		panic(err)
	}
	return analyzer
}

func (c *Configuration) MustLoadStopWords(stopWordsBytes []byte) stop_words_filter.StopWordsMap {
	rv := stop_words_filter.NewStopWordsMap()
	err := rv.LoadBytes(stopWordsBytes)
	if err != nil {
		panic(err)
	}
	return rv
}

func NewConfiguration() *Configuration {
	return &Configuration{
		Analysis: &AnalysisConfig{
			StopTokenMaps:   make(map[string]stop_words_filter.StopWordsMap),
			CharFilters:     make(map[string]analysis.CharFilter),
			Tokenizers:      make(map[string]analysis.Tokenizer),
			TokenFilters:    make(map[string]analysis.TokenFilter),
			Analyzers:       make(map[string]*analysis.Analyzer),
			DateTimeParsers: make(map[string]analysis.DateTimeParser),
		},
		Highlight: &HighlightConfig{
			Highlighters: make(map[string]search.Highlighter),
		},
	}
}

var Config *Configuration

func init() {

	// build the default configuration
	Config = NewConfiguration()

	// register stop token maps
	Config.Analysis.StopTokenMaps["da"] = Config.MustLoadStopWords(stop_words_filter.DanishStopWords)
	Config.Analysis.StopTokenMaps["nl"] = Config.MustLoadStopWords(stop_words_filter.DutchStopWords)
	Config.Analysis.StopTokenMaps["en"] = Config.MustLoadStopWords(stop_words_filter.EnglishStopWords)
	Config.Analysis.StopTokenMaps["fi"] = Config.MustLoadStopWords(stop_words_filter.FinnishStopWords)
	Config.Analysis.StopTokenMaps["fr"] = Config.MustLoadStopWords(stop_words_filter.FrenchStopWords)
	Config.Analysis.StopTokenMaps["de"] = Config.MustLoadStopWords(stop_words_filter.GermanStopWords)
	Config.Analysis.StopTokenMaps["hu"] = Config.MustLoadStopWords(stop_words_filter.HungarianStopWords)
	Config.Analysis.StopTokenMaps["it"] = Config.MustLoadStopWords(stop_words_filter.ItalianStopWords)
	Config.Analysis.StopTokenMaps["no"] = Config.MustLoadStopWords(stop_words_filter.NorwegianStopWords)
	Config.Analysis.StopTokenMaps["pt"] = Config.MustLoadStopWords(stop_words_filter.PortugueseStopWords)
	Config.Analysis.StopTokenMaps["ro"] = Config.MustLoadStopWords(stop_words_filter.RomanianStopWords)
	Config.Analysis.StopTokenMaps["ru"] = Config.MustLoadStopWords(stop_words_filter.RussianStopWords)
	Config.Analysis.StopTokenMaps["es"] = Config.MustLoadStopWords(stop_words_filter.SpanishStopWords)
	Config.Analysis.StopTokenMaps["sv"] = Config.MustLoadStopWords(stop_words_filter.SwedishStopWords)
	Config.Analysis.StopTokenMaps["tr"] = Config.MustLoadStopWords(stop_words_filter.TurkishStopWords)
	Config.Analysis.StopTokenMaps["ar"] = Config.MustLoadStopWords(stop_words_filter.ArabicStopWords)
	Config.Analysis.StopTokenMaps["hy"] = Config.MustLoadStopWords(stop_words_filter.ArmenianStopWords)
	Config.Analysis.StopTokenMaps["eu"] = Config.MustLoadStopWords(stop_words_filter.BasqueStopWords)
	Config.Analysis.StopTokenMaps["bg"] = Config.MustLoadStopWords(stop_words_filter.BulgarianStopWords)
	Config.Analysis.StopTokenMaps["ca"] = Config.MustLoadStopWords(stop_words_filter.CatalanStopWords)
	Config.Analysis.StopTokenMaps["gl"] = Config.MustLoadStopWords(stop_words_filter.GalicianStopWords)
	Config.Analysis.StopTokenMaps["el"] = Config.MustLoadStopWords(stop_words_filter.GreekStopWords)
	Config.Analysis.StopTokenMaps["hi"] = Config.MustLoadStopWords(stop_words_filter.HindiStopWords)
	Config.Analysis.StopTokenMaps["id"] = Config.MustLoadStopWords(stop_words_filter.IndonesianStopWords)
	Config.Analysis.StopTokenMaps["ga"] = Config.MustLoadStopWords(stop_words_filter.IrishStopWords)
	Config.Analysis.StopTokenMaps["fa"] = Config.MustLoadStopWords(stop_words_filter.PersianStopWords)
	Config.Analysis.StopTokenMaps["ckb"] = Config.MustLoadStopWords(stop_words_filter.SoraniStopWords)
	Config.Analysis.StopTokenMaps["th"] = Config.MustLoadStopWords(stop_words_filter.ThaiStopWords)

	// register char filters
	htmlCharFilterRegexp := regexp.MustCompile(`</?[!\w]+((\s+\w+(\s*=\s*(?:".*?"|'.*?'|[^'">\s]+))?)+\s*|\s*)/?>`)
	htmlCharFilter := regexp_char_filter.NewRegexpCharFilter(htmlCharFilterRegexp, []byte{' '})
	Config.Analysis.CharFilters["html"] = htmlCharFilter

	// register tokenizers
	whitespaceTokenizerRegexp := regexp.MustCompile(`\w+`)
	Config.Analysis.Tokenizers["single"] = single_token.NewSingleTokenTokenizer()
	Config.Analysis.Tokenizers["unicode"] = unicode_word_boundary.NewUnicodeWordBoundaryTokenizer()
	Config.Analysis.Tokenizers["unicode_th"] = unicode_word_boundary.NewUnicodeWordBoundaryCustomLocaleTokenizer("th_TH")
	Config.Analysis.Tokenizers["whitespace"] = regexp_tokenizer.NewRegexpTokenizer(whitespaceTokenizerRegexp)

	// register token filters
	Config.Analysis.TokenFilters["detect_lang"] = cld2.NewCld2Filter()
	Config.Analysis.TokenFilters["short"] = length_filter.NewLengthFilter(3, -1)
	Config.Analysis.TokenFilters["long"] = length_filter.NewLengthFilter(-1, 255)
	Config.Analysis.TokenFilters["to_lower"] = lower_case_filter.NewLowerCaseFilter()
	Config.Analysis.TokenFilters["stemmer_da"] = stemmer_filter.MustNewStemmerFilter("danish")
	Config.Analysis.TokenFilters["stemmer_nl"] = stemmer_filter.MustNewStemmerFilter("dutch")
	Config.Analysis.TokenFilters["stemmer_en"] = stemmer_filter.MustNewStemmerFilter("english")
	Config.Analysis.TokenFilters["stemmer_fi"] = stemmer_filter.MustNewStemmerFilter("finnish")
	Config.Analysis.TokenFilters["stemmer_fr"] = stemmer_filter.MustNewStemmerFilter("french")
	Config.Analysis.TokenFilters["stemmer_de"] = stemmer_filter.MustNewStemmerFilter("german")
	Config.Analysis.TokenFilters["stemmer_hu"] = stemmer_filter.MustNewStemmerFilter("hungarian")
	Config.Analysis.TokenFilters["stemmer_it"] = stemmer_filter.MustNewStemmerFilter("italian")
	Config.Analysis.TokenFilters["stemmer_no"] = stemmer_filter.MustNewStemmerFilter("norwegian")
	Config.Analysis.TokenFilters["stemmer_porter"] = stemmer_filter.MustNewStemmerFilter("porter")
	Config.Analysis.TokenFilters["stemmer_pt"] = stemmer_filter.MustNewStemmerFilter("portuguese")
	Config.Analysis.TokenFilters["stemmer_ro"] = stemmer_filter.MustNewStemmerFilter("romanian")
	Config.Analysis.TokenFilters["stemmer_ru"] = stemmer_filter.MustNewStemmerFilter("russian")
	Config.Analysis.TokenFilters["stemmer_es"] = stemmer_filter.MustNewStemmerFilter("spanish")
	Config.Analysis.TokenFilters["stemmer_sv"] = stemmer_filter.MustNewStemmerFilter("swedish")
	Config.Analysis.TokenFilters["stemmer_tr"] = stemmer_filter.MustNewStemmerFilter("turkish")

	Config.Analysis.TokenFilters["stop_token_da"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["da"])
	Config.Analysis.TokenFilters["stop_token_nl"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["nl"])
	Config.Analysis.TokenFilters["stop_token_en"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["en"])
	Config.Analysis.TokenFilters["stop_token_fi"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["fi"])
	Config.Analysis.TokenFilters["stop_token_fr"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["fr"])
	Config.Analysis.TokenFilters["stop_token_de"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["de"])
	Config.Analysis.TokenFilters["stop_token_hu"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["hu"])
	Config.Analysis.TokenFilters["stop_token_it"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["it"])
	Config.Analysis.TokenFilters["stop_token_no"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["no"])
	Config.Analysis.TokenFilters["stop_token_pt"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["pt"])
	Config.Analysis.TokenFilters["stop_token_ro"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ro"])
	Config.Analysis.TokenFilters["stop_token_ru"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ru"])
	Config.Analysis.TokenFilters["stop_token_es"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["es"])
	Config.Analysis.TokenFilters["stop_token_sv"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["sv"])
	Config.Analysis.TokenFilters["stop_token_tr"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["tr"])
	Config.Analysis.TokenFilters["stop_token_ar"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ar"])
	Config.Analysis.TokenFilters["stop_token_hy"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["hy"])
	Config.Analysis.TokenFilters["stop_token_eu"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["eu"])
	Config.Analysis.TokenFilters["stop_token_bg"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["bg"])
	Config.Analysis.TokenFilters["stop_token_ca"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ca"])
	Config.Analysis.TokenFilters["stop_token_gl"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["gl"])
	Config.Analysis.TokenFilters["stop_token_el"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["el"])
	Config.Analysis.TokenFilters["stop_token_hi"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["hi"])
	Config.Analysis.TokenFilters["stop_token_id"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["id"])
	Config.Analysis.TokenFilters["stop_token_ga"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ga"])
	Config.Analysis.TokenFilters["stop_token_fa"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["fa"])
	Config.Analysis.TokenFilters["stop_token_ckb"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["ckb"])
	Config.Analysis.TokenFilters["stop_token_th"] = stop_words_filter.NewStopWordsFilter(
		Config.Analysis.StopTokenMaps["th"])

	// register analyzers
	keywordAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "single", []string{})
	Config.Analysis.Analyzers["keyword"] = keywordAnalyzer
	simpleAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "whitespace", []string{"to_lower"})
	Config.Analysis.Analyzers["simple"] = simpleAnalyzer
	standardAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "whitespace", []string{"to_lower", "stop_token_en"})
	Config.Analysis.Analyzers["standard"] = standardAnalyzer
	detectLangAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "single", []string{"to_lower", "detect_lang"})
	Config.Analysis.Analyzers["detect_lang"] = detectLangAnalyzer

	// language specific analyzers
	danishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_da", "stemmer_da"})
	Config.Analysis.Analyzers["da"] = danishAnalyzer
	dutchAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_nl", "stemmer_nl"})
	Config.Analysis.Analyzers["nl"] = dutchAnalyzer
	englishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_en", "stemmer_en"})
	Config.Analysis.Analyzers["en"] = englishAnalyzer
	finnishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_fi", "stemmer_fi"})
	Config.Analysis.Analyzers["fi"] = finnishAnalyzer
	frenchAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_fr", "stemmer_fr"})
	Config.Analysis.Analyzers["fr"] = frenchAnalyzer
	germanAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_de", "stemmer_de"})
	Config.Analysis.Analyzers["de"] = germanAnalyzer
	hungarianAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_hu", "stemmer_hu"})
	Config.Analysis.Analyzers["hu"] = hungarianAnalyzer
	italianAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_it", "stemmer_it"})
	Config.Analysis.Analyzers["it"] = italianAnalyzer
	norwegianAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_no", "stemmer_no"})
	Config.Analysis.Analyzers["no"] = norwegianAnalyzer
	portugueseAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_pt", "stemmer_pt"})
	Config.Analysis.Analyzers["pt"] = portugueseAnalyzer
	romanianAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_ro", "stemmer_ro"})
	Config.Analysis.Analyzers["ro"] = romanianAnalyzer
	russianAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_ru", "stemmer_ru"})
	Config.Analysis.Analyzers["ru"] = russianAnalyzer
	spanishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_es", "stemmer_es"})
	Config.Analysis.Analyzers["es"] = spanishAnalyzer
	swedishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_sv", "stemmer_sv"})
	Config.Analysis.Analyzers["sv"] = swedishAnalyzer
	turkishAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode", []string{"to_lower", "stop_token_tr", "stemmer_tr"})
	Config.Analysis.Analyzers["tr"] = turkishAnalyzer
	thaiAnalyzer := Config.MustBuildNewAnalyzer([]string{}, "unicode_th", []string{"to_lower", "stop_token_th"})
	Config.Analysis.Analyzers["th"] = thaiAnalyzer

	// register ansi highlighter
	Config.Highlight.Highlighters["ansi"] = search.NewSimpleHighlighter()

	// register html highlighter
	htmlFormatter := search.NewHTMLFragmentFormatterCustom(`<span class="highlight">`, `</span>`)
	htmlHighlighter := search.NewSimpleHighlighter()
	htmlHighlighter.SetFragmentFormatter(htmlFormatter)
	Config.Highlight.Highlighters["html"] = htmlHighlighter

	// set the default analyzer
	simpleAnalyzerName := "simple"
	Config.DefaultAnalyzer = &simpleAnalyzerName

	// set the default highlighter
	htmlHighlighterName := "html"
	Config.DefaultHighlighter = &htmlHighlighterName

	// default CreateIfMissing to true
	Config.CreateIfMissing = true

	// set up the built-in date time formats

	rfc3339NoTimezone := "2006-01-02T15:04:05"
	rfc3339NoTimezoneNoT := "2006-01-02 15:04:05"
	rfc3339NoTime := "2006-01-02"

	Config.Analysis.DateTimeParsers["dateTimeOptional"] = flexible_go.NewFlexibleGoDateTimeParser(
		[]string{
			time.RFC3339Nano,
			time.RFC3339,
			rfc3339NoTimezone,
			rfc3339NoTimezoneNoT,
			rfc3339NoTime,
		})
	dateTimeOptionalName := "dateTimeOptional"
	Config.DefaultDateTimeFormat = &dateTimeOptionalName
}