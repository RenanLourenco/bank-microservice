Hello everyone, welcome to this project, a simulation of a banking system. The primary objective behind this endeavor is to enhance my proficiency in microservices. Here's an overview of the services I've developed:

1. **Auth:** This service handles authentication, login, registration, and token updates.
2. **Broker:** Serving as the central hub, this service receives all requests and efficiently distributes them to other services via gRPC or places them in a RabbitMQ queue.
3. **Listener:** Specifically designed to await messages originating from the RabbitMQ queue, this service focuses on transactions and deposits.
4. **Transaction:** This service encompasses all the essential methods for facilitating seamless transactions between users.

The next steps involve implementing authentication for another user profile, specifically designed for legal entities. Additionally, I plan to develop the front-end and establish automated tests, including end-to-end (e2e) testing. 
As for deployment, I am currently exploring the possibility of utilizing a Kubernetes cluster. Stay tuned for further updates!
