Instructions to use:
To run: `docker compose up`
To run tests `go test` in the same directory
Note: It can be better to run with `docker compose up --build && docker compose rm -f` to ensure a clean
build each time. This can be slow because I have not implemented exposing a build cache to the build
container.

Notes:
- I don't feel like I did a good enough job of error handling. More useful error responses would be helpful, 
  and some things are checked in the DB layer, that should return messages to the user with a 400 that should
  have been caught in input validation in the UI layer, or through structured error responses.
- This file is largely a stream of thought as I was working on the task. I lightly proofread it, but there
  may still be some inconsistencies with what I have ended up delivering. My apologies if this is the case.
- I feel like I could write another 1000 lines of test code, but I don't feel like it would be worth me 
  writing or you reviewing all of that code. Let's have a conversation about testing strategy in an
  interview instead?

Functional requirements
- Balance in an account should never be able to fall below 0 as I'm not handling pre-approved overdraft
- Transaction history should be auditable
- Currency values should be handled as integers in cents to prevent errors inherent in fractional number 
  representation in computers.
- The API service should be stateless

Things I thought about but decided not to build as they are out of scope:
- Account balance endpoint 
- Account withdrawl - needed to allow withdrawing cash
- Authentication - Only the person who opened an account should be able to withdraw money it
- Service redundancy - Run multiple copies of the API server to prevent the crash of an instance
  from causing an outage
- API gateway/load balancer - distribute traffic across different instances of the api service
- DB replication - Guaranteed data persistence requires multiple copies of the data to be stored.
  The easiest way to do this is with multiple-hot DB replication

Other design decisions
- I'm using GORM to leverage schema management. I usually prefer using a simpler query builder, and 
  managing the schema separately, as (in my experience) ORMs usually don't handle migrations very
  well, and can produce very inefficient queries. And removing the abstraction layer requires devs
  to better understand what it is they are doing when interacting with the DB resulting in higher
  quality solutions and more team members being able to handle DB related production incidents. 
  Additionally, code is cleaner when you aren't constantly constructing objects as the means of filtering
  results of queries.
- I don't explicitly do any connection pooling as Gorm abstracts connections behind a pool in it's DB 
  construct already. If I were using a simpler DB abstraction layer, I would store the pool in the server
  struct, and add a middleware to retrieve a connection/session from the pool and pass it to each handler 
  in the gin context.
- Account numbers are not usually RFC4122 UUIDs, but I don't feel like adding complexity/a dependency
  for correctly representing IBANS (or other standard) would be useful for this exercise, so long as
  account numbers are guaranteed to be unique. UUIDs were chosen as an arbitrary "unique, structured, and
  vaguely representative" solution. Incremental ints are not used as they could allow guessing the number
  someone else received/will receive, which poses a potential security issue.
- Account numbers in transactions are not linked with foreign keys to existing accounts to allow recording
  transfers to/from other banks. Ideally there would be a foreign key to represent that either the source
  or destination account exists in our account table, but I have no idea how to do that with GORM. This 
  could be enforced at the DB layer through stored procedures/functions, but I feel that is too complicated
  for this exercise.
- I'm using a DB that supports transactions to demonstrate that I understand how these should be used in 
  production systems. I considered using a simpler storage solution and treating the transaction history as
  the sole source of truth (which I would still consider true semantically for the sake of data integrity)
- I'm using the 0 UUID to represent cash interactions. In a production system it would be more appropriate
  to have an account for each ATM and branch to facilitate auditing.
- DB interactions are done without contexts in the interest of simplicity
- Using 64bit ints to avoid having to deal with overflow in calculations for any situation short of 
  Zimbabwe-style hyperinflation, at which point we have other problems to worry about.
- I am purposefully querying the accounts table before writing any transactions to use row-level locks
  as cross-process mutexes without having to lock whole tables. However, I have not checked that PG is
  using row level locks.
- I did not annotate every error with context for the sake of brevity. This would be desireable in 
  most production systems.
- I used the stdlib log package for simplicity, but it's quite rudimentary. I would use something with more
  structure in production eg: zap
- I considered doing table driven testing. But having account numbers generated on the server would mean I
  need to track multiple account numbers across calls, which would become finicky and fragile.
- Types used for business logic are defined in the API package, so that the types and their means of 
  serialisation etc can be exported for easy construction of an SDK/integrations, and for implementing
  interfaces for modular components such as storage adapters.
- The AccountStore and DB Connections are handled separately for implementing session pooling, even though
  I didn't end up needing to implement that.
- I'm limiting numerical values in the api to uint32 because I think that's plenty high enough for real
  usecases, and it helps us keep clear of overflow issues
- I chose to not mock the DB as that would take more time and add complexity. And I have found that 
  doing so tends to cause more problems than it solves in web APIs.
- lifecycle_test.go is in the main package so that I could reuse the DB setup code. Ideally tests against
  external APIs should be in their own packages.
- I query the DB directly in the tests to check the state is as intended. This is usually not desirable
  as it tests the implementation, not behaviour, can be very fragile, creates direct dependencies on parts 
  of the system under test instead of its interface, and can require updating the test to not break it 
  when making unrelated changes. However, there are no API endpoints for reading, which I would prefer to
  use when writing tests (which would need to be tested anyway).
- The error strings are ugly. I couldn't be bothered fixing it
- Tests using a live DB like this should use a separate schema for each test/suite, but I chose not to to 
  save time and reduce complexity


AI disclosure
I used claude to generate the docker and docker-compose files, and reviewed them to make sure they are
doing rougly what I want. I did not put effort into them as they serve only to facilitate using postgres
as the DB with easy development, review, and testing.
I also used claud to template a basic go service using gorm, pg, and gin. Aside from some of main.go the
code I have submitted was written entirely by me.
