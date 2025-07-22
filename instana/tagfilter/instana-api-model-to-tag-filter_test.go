package tagfilter_test

import (
	"fmt"
	"testing"

	"github.com/gessnerfl/terraform-provider-instana/utils"

	"github.com/stretchr/testify/require"

	"github.com/gessnerfl/terraform-provider-instana/instana/restapi"
	. "github.com/gessnerfl/terraform-provider-instana/instana/tagfilter"
)

const (
	invalidOperator   = "invalid operator"
	tagFilterOperator = "tag filter operator"
	tagFilterName     = "name"
)

func TestShouldMapEmptyTagFilterExpressionFromInstanaAPI(t *testing.T) {
	for _, operator := range restapi.SupportedLogicalOperatorTypes {
		t.Run(fmt.Sprintf("TestShouldMapEmpty%sTagFilterExpressionFromInstnaAPI", string(operator)), func(t *testing.T) {
			op := operator
			expression := &restapi.TagFilter{
				Type:            restapi.TagFilterExpressionType,
				LogicalOperator: &op,
			}

			runTestCaseForMappingFromAPI(expression, nil, t)
		})
	}
}

func TestShouldMapStringTagFilterFromInstanaAPI(t *testing.T) {
	value := "value"
	input := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.EqualsOperator, value)

	comparison := &ComparisonExpression{
		Entity:      &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
		Operator:    Operator(restapi.EqualsOperator),
		StringValue: &value,
	}

	testMappingOfTagFilterFromInstanaApi(input, comparison, t)
}

func TestShouldMapNumberTagFilterFromInstanaAPI(t *testing.T) {
	value := int64(1234)
	input := restapi.NewNumberTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.EqualsOperator, value)

	comparison := &ComparisonExpression{
		Entity:      &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
		Operator:    Operator(restapi.EqualsOperator),
		NumberValue: &value,
	}

	testMappingOfTagFilterFromInstanaApi(input, comparison, t)
}

func TestShouldMapBooleanTagFilterFromInstanaAPI(t *testing.T) {
	value := true
	input := restapi.NewBooleanTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.EqualsOperator, value)

	comparison := &ComparisonExpression{
		Entity:       &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
		Operator:     Operator(restapi.EqualsOperator),
		BooleanValue: &value,
	}

	testMappingOfTagFilterFromInstanaApi(input, comparison, t)
}

func TestShouldMapComparisonTagFilterWithTagKeyValueFromInstanaAPI(t *testing.T) {
	key := "key"
	value := "value"
	input := restapi.NewTagTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.EqualsOperator, key, value)

	comparison := &ComparisonExpression{
		Entity:      &EntitySpec{Identifier: tagFilterName, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
		Operator:    Operator(restapi.EqualsOperator),
		StringValue: &value,
	}

	testMappingOfTagFilterFromInstanaApi(input, comparison, t)
}

func testMappingOfTagFilterFromInstanaApi(tagFilter *restapi.TagFilter, comparison *ComparisonExpression, t *testing.T) {
	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{Comparison: comparison},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(tagFilter, expectedResult, t)
}

func TestShouldMapAllSupportedComparisonOperatorsFromInstanaAPI(t *testing.T) {
	for _, v := range restapi.SupportedComparisonOperators {
		t.Run(fmt.Sprintf("test mapping of %s", v), testMappingOfSupportedComparisonOperatorsFromInstanaAPI(v))
	}
}

func testMappingOfSupportedComparisonOperatorsFromInstanaAPI(operator restapi.ExpressionOperator) func(t *testing.T) {
	return func(t *testing.T) {
		value := "value"
		input := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, tagFilterName, operator, value)

		expectedResult := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							Comparison: &ComparisonExpression{
								Entity:      &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator:    Operator(operator),
								StringValue: &value,
							},
						},
					},
				},
			},
		}

		runTestCaseForMappingFromAPI(input, expectedResult, t)
	}
}

func TestShouldFailToMapTagFilterFromInstanaAPIWhenOperatorIsNotSupported(t *testing.T) {
	value := "value"
	input := restapi.NewStringTagFilter(restapi.TagFilterEntityDestination, tagFilterName, "FOO", value)

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), invalidOperator)
	require.Contains(t, err.Error(), tagFilterOperator)
}

func TestShouldMapAllSupportedUnaryOperationsFromInstanaAPI(t *testing.T) {
	for _, v := range restapi.SupportedUnaryExpressionOperators {
		t.Run(fmt.Sprintf("test mapping of %s ", v), testMappingOfSupportedUnaryOperationFromInstanaAPI(v))
	}
}

func testMappingOfSupportedUnaryOperationFromInstanaAPI(operator restapi.ExpressionOperator) func(t *testing.T) {
	return func(t *testing.T) {
		input := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, operator)

		expectedResult := &FilterExpression{
			Expression: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: Operator(operator),
							},
						},
					},
				},
			},
		}

		runTestCaseForMappingFromAPI(input, expectedResult, t)
	}
}

func TestShouldMapUnaryTagFilterWithTagKeyFromInstanaAPI(t *testing.T) {
	key := "key"
	input := restapi.NewUnaryTagFilterWithTagKey(restapi.TagFilterEntityDestination, tagFilterName, &key, restapi.NotEmptyOperator)

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						UnaryOperation: &UnaryOperationExpression{
							Entity:   &EntitySpec{Identifier: tagFilterName, TagKey: &key, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator: Operator(restapi.NotEmptyOperator),
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldFailToMapTagFilterFromInstanaAPIWhenUnaryOperationIsNotSupported(t *testing.T) {
	input := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, "FOO")

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), invalidOperator)
	require.Contains(t, err.Error(), tagFilterOperator)
}

func TestShouldFailToMapTagFilterExpressionElementFromInstanaAPIWhenTypeIsMissing(t *testing.T) {
	name := tagFilterName
	operator := restapi.ExpressionOperator("FOO")
	input := &restapi.TagFilter{
		Name:     &name,
		Operator: &operator,
	}

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "unsupported tag filter expression")
}

func TestShouldMapLogicalAndWithTwoPrimaryExpressionsFromInstanaAPI(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	and := Operator(restapi.LogicalAnd)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name1", restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	input := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression1, primaryExpression2})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: "name1", Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
							Operator: &and,
							Right: &LogicalAndExpression{
								Left: &BracketExpression{
									Primary: &PrimaryExpression{
										UnaryOperation: &UnaryOperationExpression{
											Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
											Operator: operator,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalAndWithThreePrimaryExpressionsFromInstanaAPI(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	and := Operator(restapi.LogicalAnd)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name1", restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	primaryExpression3 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name3", restapi.IsEmptyOperator)
	input := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression1, primaryExpression2, primaryExpression3})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: "name1", Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
							Operator: &and,
							Right: &LogicalAndExpression{
								Left: &BracketExpression{
									Primary: &PrimaryExpression{
										UnaryOperation: &UnaryOperationExpression{
											Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
											Operator: operator,
										},
									},
								},
								Operator: &and,
								Right: &LogicalAndExpression{
									Left: &BracketExpression{
										Primary: &PrimaryExpression{
											UnaryOperation: &UnaryOperationExpression{
												Entity:   &EntitySpec{Identifier: "name3", Origin: utils.StringPtr(EntityOriginDestination.Key())},
												Operator: operator,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalAndWithTwoElementsFromInstanaAPIWhereTheFirstElementIsAPrimaryExpressionAndTheSecondElementIsAnotherLogicalAnd(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	and := Operator(restapi.LogicalAnd)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	nestedAnd := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression2, primaryExpression2})
	input := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression1, nestedAnd})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
							Operator: &and,
							Right: &LogicalAndExpression{
								Left: &BracketExpression{
									Bracket: &LogicalOrExpression{
										Left: &LogicalAndExpression{
											Left: &BracketExpression{
												Primary: &PrimaryExpression{
													UnaryOperation: &UnaryOperationExpression{
														Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
														Operator: operator,
													},
												},
											},
											Operator: &and,
											Right: &LogicalAndExpression{
												Left: &BracketExpression{
													Primary: &PrimaryExpression{
														UnaryOperation: &UnaryOperationExpression{
															Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
															Operator: operator,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalAndWithTwoElementsFromInstanaAPIWhereTheFirstElementIsAPrimaryExpressionAndTheSecondElementIsALogicalOr(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	and := Operator(restapi.LogicalAnd)
	or := Operator(restapi.LogicalOr)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	nestedOr := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression2, primaryExpression2})
	input := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression1, nestedOr})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
							Operator: &and,
							Right: &LogicalAndExpression{
								Left: &BracketExpression{
									Bracket: &LogicalOrExpression{
										Left: &LogicalAndExpression{
											Left: &BracketExpression{
												Primary: &PrimaryExpression{
													UnaryOperation: &UnaryOperationExpression{
														Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
														Operator: operator,
													},
												},
											},
										},
										Operator: &or,
										Right: &LogicalOrExpression{
											Left: &LogicalAndExpression{
												Left: &BracketExpression{
													Primary: &PrimaryExpression{
														UnaryOperation: &UnaryOperationExpression{
															Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
															Operator: operator,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldUnwrapLogicalAndFromInstanaAPIWhenOnlyOneElementIsProvided(t *testing.T) {
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	input := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						UnaryOperation: &UnaryOperationExpression{
							Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator: Operator(restapi.IsEmptyOperator),
						},
					},
				},
			},
		},
	}

	mapper := NewMapper()
	result, err := mapper.FromAPIModel(input)

	require.NoError(t, err)
	require.Equal(t, expectedResult, result)
}

func TestShouldMapLogicalOrWithTwoPrimaryExpressionsFromInstanaAPI(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	or := Operator(restapi.LogicalOr)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name1", restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression1, primaryExpression2})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: "name1", Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
						},
						Operator: &or,
						Right: &LogicalOrExpression{
							Left: &LogicalAndExpression{
								Left: &BracketExpression{
									Primary: &PrimaryExpression{
										UnaryOperation: &UnaryOperationExpression{
											Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
											Operator: operator,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalOrWithThreePrimaryExpressionsFromInstanaAPI(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	or := Operator(restapi.LogicalOr)
	primaryExpression1 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name1", restapi.IsEmptyOperator)
	primaryExpression2 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name2", restapi.IsEmptyOperator)
	primaryExpression3 := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, "name3", restapi.IsEmptyOperator)
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression1, primaryExpression2, primaryExpression3})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						UnaryOperation: &UnaryOperationExpression{
							Entity:   &EntitySpec{Identifier: "name1", Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator: operator,
						},
					},
				},
			},
			Operator: &or,
			Right: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: "name2", Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: operator,
							},
						},
					},
				},
				Operator: &or,
				Right: &LogicalOrExpression{
					Left: &LogicalAndExpression{
						Left: &BracketExpression{
							Primary: &PrimaryExpression{
								UnaryOperation: &UnaryOperationExpression{
									Entity:   &EntitySpec{Identifier: "name3", Origin: utils.StringPtr(EntityOriginDestination.Key())},
									Operator: operator,
								},
							},
						},
					},
				},
			},
		}}}},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalOrWithTwoElementsFromInstanaAPIWhereFirstElementIsALogicalAndAndTheOtherElementIsPrimaryExpression(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	or := Operator(restapi.LogicalOr)
	and := Operator(restapi.LogicalAnd)
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	nestedAnd := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression, primaryExpression})
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{nestedAnd, primaryExpression})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Bracket: &LogicalOrExpression{
						Left: &LogicalAndExpression{
							Left: &BracketExpression{
								Primary: &PrimaryExpression{
									UnaryOperation: &UnaryOperationExpression{
										Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
										Operator: operator,
									},
								},
							},
							Operator: &and,
							Right: &LogicalAndExpression{
								Left: &BracketExpression{
									Primary: &PrimaryExpression{
										UnaryOperation: &UnaryOperationExpression{
											Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
											Operator: operator,
										},
									},
								},
							},
						},
					},
				},
			},
			Operator: &or,
			Right: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: operator,
							},
						},
					},
				},
			},
		}}}},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalOrWithTwoElementsFromInstanaAPIWhereFirstElementIsAPrimaryExpressionAndTheOtherElementIsALogicalOr(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	or := Operator(restapi.LogicalOr)
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	nestedOr := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression, primaryExpression})
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression, nestedOr})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						UnaryOperation: &UnaryOperationExpression{
							Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator: operator,
						},
					},
				},
			},
			Operator: &or,
			Right: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: operator,
							},
						},
					},
				},
				Operator: &or,
				Right: &LogicalOrExpression{
					Left: &LogicalAndExpression{
						Left: &BracketExpression{
							Primary: &PrimaryExpression{
								UnaryOperation: &UnaryOperationExpression{
									Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
									Operator: operator,
								},
							},
						},
					},
				},
			}}}},
		}}}},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldMapLogicalOrWithTwoElementsWhereFirstElementIsAPrimaryExpressionAndTheOtherElementIsALogicalAnd(t *testing.T) {
	operator := Operator(restapi.IsEmptyOperator)
	or := Operator(restapi.LogicalOr)
	and := Operator(restapi.LogicalAnd)
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	nestedAnd := restapi.NewLogicalAndTagFilter([]*restapi.TagFilter{primaryExpression, primaryExpression})
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression, nestedAnd})

	expectedResult := &FilterExpression{
		Expression: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
			Left: &LogicalAndExpression{
				Left: &BracketExpression{
					Primary: &PrimaryExpression{
						UnaryOperation: &UnaryOperationExpression{
							Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
							Operator: operator,
						},
					},
				},
			},
			Operator: &or,
			Right: &LogicalOrExpression{Left: &LogicalAndExpression{Left: &BracketExpression{Bracket: &LogicalOrExpression{
				Left: &LogicalAndExpression{
					Left: &BracketExpression{
						Primary: &PrimaryExpression{
							UnaryOperation: &UnaryOperationExpression{
								Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
								Operator: operator,
							},
						},
					},
					Operator: &and,
					Right: &LogicalAndExpression{
						Left: &BracketExpression{
							Primary: &PrimaryExpression{
								UnaryOperation: &UnaryOperationExpression{
									Entity:   &EntitySpec{Identifier: tagFilterName, Origin: utils.StringPtr(EntityOriginDestination.Key())},
									Operator: operator,
								},
							},
						},
					},
				},
			}}}},
		}}}},
	}

	runTestCaseForMappingFromAPI(input, expectedResult, t)
}

func TestShouldFailToMapLogicalOrWhenFirstElementIsALogicalOrExpression(t *testing.T) {
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	nestedOr := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression, primaryExpression})
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{nestedOr, primaryExpression})

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "logical or is not allowed for first element")
}

func TestShouldFailToMapLogicalOrWhenOnlyOneElementIsProvided(t *testing.T) {
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	input := restapi.NewLogicalOrTagFilter([]*restapi.TagFilter{primaryExpression})

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "at least two elements are expected for logical or")
}

func TestShouldFailToMapTagFilterExpressionFromInstanaAPIWhenLogicalOperatorIsNotValid(t *testing.T) {
	primaryExpression := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
	operator := restapi.LogicalOperatorType("FOO")
	input := &restapi.TagFilter{
		Type:            restapi.TagFilterExpressionType,
		LogicalOperator: &operator,
		Elements:        []*restapi.TagFilter{primaryExpression, primaryExpression},
	}

	mapper := NewMapper()
	_, err := mapper.FromAPIModel(input)

	require.NotNil(t, err)
	require.Contains(t, err.Error(), "invalid logical operator")
}

func TestShouldReturnMappingErrorWhenAnyElementOfTagFilterExpressionIsNotValid(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Run(fmt.Sprintf("TestShouldReturnMappingErrorWhenElement%dOfTagFilterExpressionIsNotValid", i), func(t *testing.T) {
			invalidElement := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, "INVALID")
			validElement := restapi.NewUnaryTagFilter(restapi.TagFilterEntityDestination, tagFilterName, restapi.IsEmptyOperator)
			elements := make([]*restapi.TagFilter, 5)
			for j := 0; j < 5; j++ {
				if j == i {
					elements[j] = invalidElement
				} else {
					elements[j] = validElement
				}
			}
			input := restapi.NewLogicalOrTagFilter(elements)

			mapper := NewMapper()
			_, err := mapper.FromAPIModel(input)

			require.NotNil(t, err)
			require.Contains(t, err.Error(), invalidOperator)
			require.Contains(t, err.Error(), tagFilterOperator)
		})
	}
}

func runTestCaseForMappingFromAPI(input *restapi.TagFilter, expectedResult *FilterExpression, t *testing.T) {
	mapper := NewMapper()
	result, err := mapper.FromAPIModel(input)

	require.Nil(t, err)
	require.Equal(t, expectedResult, result)
}
