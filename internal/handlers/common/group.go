// Copyright 2021 FerretDB Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"strings"

	"github.com/FerretDB/FerretDB/internal/types"
	"github.com/FerretDB/FerretDB/internal/util/must"
	"golang.org/x/exp/maps"
)

type GroupContext struct {
	parents []string
	fields  map[string]interface{}
	groups  []string
}

func NewGroupContext() GroupContext {
	ctx := GroupContext{
		parents: []string{},
		fields:  map[string]interface{}{},
		groups:  []string{},
	}

	return ctx
}

func (c *GroupContext) AddField(name string, value interface{}) {
	c.fields[name] = value
}

func (c *GroupContext) AddGroup(name string) {
	c.groups = append(c.groups, name)
}

func (c *GroupContext) GetParent() string {
	return c.parents[len(c.parents)-1]
}

func (c *GroupContext) FieldAsString() string {
	str := "json_build_object('$k', jsonb_build_array("

	for _, key := range maps.Keys(c.fields) {
		str += fmt.Sprintf("'%s', ", key)
	}

	str = strings.TrimSuffix(str, ", ") + "), "

	for key, value := range c.fields {
		str += fmt.Sprintf("'%s', %v, ", key, value)
	}
	str = strings.TrimSuffix(str, ", ") + ") AS _jsonb"

	return str
}

func ParseGroup(ctx *GroupContext, key string, value interface{}) error {
	switch key {
	case "_id":
		ctx.AddField("_id", value)

	case "$count":
		ctx.AddField(ctx.GetParent(), "COUNT(*)")

	default:
		if strings.HasPrefix(key, "$") {
			return NewWriteErrorMsg(
				ErrFailedToParse,
				fmt.Sprintf(
					"Unknown top level operator: %s. Expected a valid aggregate modifier", key,
				),
			)
		}

		switch v := value.(type) {
		case *types.Document:
			if key != "" {
				ctx.parents = append(ctx.parents, key)
			}
			for _, key := range v.Keys() {
				value := must.NotFail(v.Get(key))
				err := ParseGroup(ctx, key, value)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func AggregateGroup(group *types.Document) (*string, *string, error) {
	ctx := NewGroupContext()

	err := ParseGroup(&ctx, "", group)
	if err != nil {
		return nil, nil, err
	}

	fields := ctx.FieldAsString()
	groups := ""

	return &fields, &groups, nil
}
