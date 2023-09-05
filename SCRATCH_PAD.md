


## pulling out voting records from:

https://members-api.parliament.uk/index.html


find the person like:
```
curl -X 'GET' \
  'https://members-api.parliament.uk/api/Members/Search?Name=Paul%20Holmes&skip=0&take=20' \
  -H 'accept: text/plain'
```

to get their ID: items[x].value.id
```
{
  "items": [
    {
      "value": {
        "id": 1404,
        "nameListAs": "Holmes, Paul",
```


## to get all votes

```
curl -X 'GET' \
  'https://members-api.parliament.uk/api/Members/4803/Voting?house=1&page=1' \
  -H 'accept: text/plain'
```
- gets all the votes this MP has participated on (possibly not necessary tbh)



get all votes for a particular motion:
https://commonsvotes-api.parliament.uk/data/division/1617.json




## policy interests
TODO - how to categorise them???
- was this done by they work for you??? TODO figure out


https://twitter.com/intent/tweet?original_referer=https%3A%2F%2Fwww.theyworkforyou.com%2F&ref_src=twsrc%5Etfw%7Ctwcamp%5Ebuttonembed%7Ctwterm%5Eshare%7Ctwgr%5E&text=Voting%20record%20-%20Paul%20Holmes%20MP%2C%20Eastleigh%20-%20TheyWorkForYou&url=https%3A%2F%2Fwww.theyworkforyou.com%2Fmp%2F25808%2Fpaul_holmes%2Feastleigh%2Fvotes%3Fpolicy%3Dmisc
translates to: https://www.theyworkforyou.com/mp/25808/paul_holmes/eastleigh/votes?policy=misc 
top level view is: https://www.theyworkforyou.com/mp/25808/paul_holmes/eastleigh/votes?policy=


### get the list of policy interests with: 
```
curl -X 'GET' \
  'https://members-api.parliament.uk/api/Reference/PolicyInterests' \
  -H 'accept: text/plain'
what to use this with?
```



this looks promising:
https://www.publicwhip.org.uk/mp.php?mpid=42172&dmp=6703

`Paul Holmes MP, Eastleigh 

voted strongly against the policy

Human Rights and Equality`

seems to be based around `policies`:
https://www.publicwhip.org.uk/faq.php#policies

