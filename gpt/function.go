package gpt

/*
	function = {
	   "name": "predict_area_of_learning",
	   "description": "Predict the EYFS area of learning for a given text",
	   "parameters": {
	       "type": "object",
	       "properties": {
	           "prediction": {
	               "type": "array",
	               "items": {
	                   "type": "string",
	                   "enum": [
	                       "Communication and Language",
	                       "Personal, Social and Emotional Development",
	                       "Physical Development",
	                       "Literacy",
	                       "Mathematics",
	                       "Understanding the World",
	                       "Expressive Arts and Design",
	                       "None"
	                   ]
	               },
	               "description": "The predicted areas of learning."
	           }
	       },
	       "required": [
	           "prediction"
	       ]
	   }
	}
*/
type Function struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  FunctionParams `json:"parameters"`
}

type FunctionParams struct {
	Type       string                        `json:"type"`
	Properties map[string]FunctionPrediction `json:"properties"`
	Required   []string                      `json:"required"`
}

type FunctionPrediction struct {
	Type        string        `json:"type"`
	Items       FunctionItems `json:"items"`
	Description string        `json:"description"`
}

type FunctionItems struct {
	Type string   `json:"type"`
	Enum []string `json:"enum"`
}
