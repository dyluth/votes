# Votes


this go project is designed to:
1) take messages that an MP might write
2) classify those messages into the different policy groups (defined by www.publicwhip.org.uk) usoing chatGPT
3) find the MP's voting record from www.publicwhip.org.uk 
4) build some sort of response highlighing their voting position on the related policy to post as a reply.

## setup
you will need: 
* an openAI API account with money in it.
* a twitter BASIC account - currently $100/month :'( 

run `twitter-auth-setup.py` - this will print out a URL, have the twitter user we are going to control visit that URL and grant access, then provide the code back to the script.
the script will then print out the users key and secret.

export those as `TWITTER_USER_TOKEN` and `TWITTER_USER_SECRET` so the main project can act as that user.



## running the project
make sure to export the following env variables:

* `GPT_API_KEY` - API Key to access chat GPT
* `OPENAI_API_KEY` - duplicate of `GPT_API_KEY` - hardcoded in one of the libraries we use

the following are provided by twitter 
`TWITTER_API_KEY` 
`TWITTER_API_KEY_SECRET`
`TWITTER_BEARER_TOKEN`
`TWITTER_ACCESS_TOKEN`
`TWITTER_ACCESS_TOKEN_SECRET`

the following are provided by `twitter-auth-setup.py` for the user we are going to link to
* `TWITTER_USER_TOKEN` 
* `TWITTER_USER_SECRET` 

run the program with:
`go run cmd/twitter/main.go`



# examples for how the sub parts work
`go run cmd/withgpt/main.go` will give you an idea of how it's supposed to work.
`go run cmd/service/main.go` runs as a service that hosts an endpoint you can send requests to (see its readme)

`go run cmd/service/main.go` - the classifier was originally designed to be its own microservice, but due to time constraints was redesigned as a monolith

# different parts of the project
## data from PublicWhip
this unfortunately has to be scraped from their website
* that is largely what the `publicwhip` package does
however - getting a list of MPs along with their internal ID number takes time (and wont change often), so is currently cached in the file `mpData` - if you delete this file, next time it runs it will be regenerated. 


## members-api.parliament.uk
there is code that can talk to the official parliament api, but so far we have not needed this

## gpt
this is the package that talks to chatGPT and uses it to classify mesages using the policies from publicwhip
NOTE: we are using GetReducedPolicies to reduce our GPT prompt size (and cost!) - just using a hardcoded subset of interesting policies rather than all of them - this can be modified in the `got/completion.go` code 
