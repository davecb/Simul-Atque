# Simul Atque
Latin for "as soon as", this is a web server that
waits a tenth of a second and then just returns.
It's a convenient dummy service with an adjustable
number of servers and service time 

As soon as the service time is over, each server returns
0 bytes. 

The service time and number of servers can be set with 
```
 --servers N 
 --service-time T
 ```
 where T is time in milliseconds.
