# Decision Log Doc

This document is used to record the context and justification for the main decisions made over the course of the project. There are many choices to be made in a project and this log can serve as a useful reference in the future. The explanations are not exhaustive, but list the main things that were considered at the time.


# Format Example

Each documented decision can roughly follow this format (as is applicable)

## Topic
_decision:_  ...

_alternatives:_ ...  

_justification:_ ...

_potential downsides:_ ... 


# Decisions

## Programming Language
_decision:_   [__GO__](https://go.dev/)  

_alternatives:_    python, ruby, C, Javascript  

_justification:_ go is a modern statically typed language that is highly performant, easy to learn, and designed with maintenance and collaboration in mind. It is growing in popularity for web service and API developers. As there is not a strong existing development culture in the organisation around one language, it seemed worthwhile to evaluate how quickly we can learn and become productive using this new language.

_potential downsides:_


## Framework
_decision:_ [__GIN__](https://github.com/gin-gonic/gin)  

_alternatives:_  no framework,   

_justification:_ one of the fastest and most popular 
frameworks for http services with lots of support online. 


## Architecture Pattern
_decision:_ __asynchronous task queues__ (using Redis Lists)   

_alternatives:_  synchronous task handling, Stream-based asynchronous communication

_justification:_ improves scalability, even if request rates are high or jobs take a long time, the jobs will still complete and it just takes longer to receive the download email.   
Redis is a highly performant in memory database with free to use container images. 

_potential downsides:_  more complicated than synchronous requests

_background info:_

* [Redis blog: What to Choose for Your Synchronous and Asynchronous Communication Needs](https://redis.com/blog/what-to-choose-for-your-synchronous-and-asynchronous-communication-needs-redis-streams-redis-pub-sub-kafka-etc-best-approaches-synchronous-asynchronous-communication/)



## Hosting
_decision:_ __Google Cloud Services App Engine__  (temporary, IT preference is to eventually host on-prem running in containers) 

_alternatives:_  Google Cloud Run, Cloud Functions, Digital Ocean droplet, VM's, on-premise etc. ... 

_justification:_ App engine provides integrated Continuous Deployment and SLL certificates, reasonable free tier, well integrated with Google Cloud Buckets.  

_potential downsides:_ cost (there are cheaper alternatives), Google Cloud requires some familiarisation time



## Database
_decision:_  TBD

_alternatives:_ ...  

_justification:_ ...

_potential downsides:_ 
