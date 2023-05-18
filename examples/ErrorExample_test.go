//  Copyright hyperjumptech/grule-rule-engine Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package examples

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/farisekananda/grule-rule-engine/ast"
	"github.com/farisekananda/grule-rule-engine/builder"
	"github.com/farisekananda/grule-rule-engine/engine"
	"github.com/farisekananda/grule-rule-engine/pkg"
	"github.com/stretchr/testify/assert"
)

type MyFact2 struct {
	IntAttribute     int64
	StringAttribute  string
	BooleanAttribute bool
	FloatAttribute   float64
	TimeAttribute    time.Time
	WhatToSay        string
}

func (mf *MyFact2) GetWhatToSay(sentence string) string {
	return fmt.Sprintf("Let say \"%s\"", sentence)
}

type Repo struct{}

func (r *Repo) GetSentence() (string, error) {
	return "Hello from the repo", nil
}

func (r *Repo) GetSentenceWithError() (string, error) {
	return "This will not shown", errDefined
}

var errDefined = errors.New("unfortunately this returns an error")

func TestError(t *testing.T) {
	successFact := MyFact2{
		IntAttribute:     123,
		StringAttribute:  "Some string value",
		BooleanAttribute: true,
		FloatAttribute:   1.234,
		TimeAttribute:    time.Now(),
	}
	dataCtx := ast.NewDataContext()
	err := dataCtx.Add("MF", &successFact)
	assert.NoError(t, err)

	repo := Repo{}

	err = dataCtx.Add("R", &repo)
	assert.NoError(t, err)

	// Prepare knowledgebase library and load it with our rule.
	knowledgeLibrary := ast.NewKnowledgeLibrary()
	ruleBuilder := builder.NewRuleBuilder(knowledgeLibrary)

	drls := `
rule CheckValues "Check the default values" salience 10 {
    when 
        MF.IntAttribute == 123 && MF.StringAttribute == "Some string value"
    then
        Sentence = R.GetSentence();
        MF.WhatToSay = MF.GetWhatToSay(Sentence);
		Retract("CheckValues");
}

rule CheckValuesError "Check the default values with error" salience 10 {
    when 
        MF.IntAttribute == 1234 && MF.StringAttribute == "Some string value"
    then
        Sentence = R.GetSentenceWithError();
        MF.WhatToSay = MF.GetWhatToSay(Sentence);
		Retract("CheckValuesError");
}
`
	byteArr := pkg.NewBytesResource([]byte(drls))
	err = ruleBuilder.BuildRuleFromResource("Tutorial", "0.0.1", byteArr)
	assert.NoError(t, err)

	knowledgeBase := knowledgeLibrary.NewKnowledgeBaseInstance("Tutorial", "0.0.1")

	engine := engine.NewGruleEngine()
	err = engine.Execute(dataCtx, knowledgeBase)
	assert.NoError(t, err)
	assert.Equal(t, "Let say \"Hello from the repo\"", successFact.WhatToSay)
	println(successFact.WhatToSay)

	errFact := MyFact2{
		IntAttribute:     1234,
		StringAttribute:  "Some string value",
		BooleanAttribute: true,
		FloatAttribute:   1.234,
		TimeAttribute:    time.Now(),
	}

	dataCtx2 := ast.NewDataContext()
	err = dataCtx2.Add("MF", &errFact)
	assert.NoError(t, err)

	err = dataCtx2.Add("R", &repo)
	assert.NoError(t, err)

	err = engine.Execute(dataCtx2, knowledgeBase)
	assert.ErrorIs(t, err, errDefined)
	assert.Equal(t, "", errFact.WhatToSay)
	println(errFact.WhatToSay)
}
