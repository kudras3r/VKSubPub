# VKSubPub

## Description

There is a publisher-subscriber message worker as gRPC service.

## About 

VKSubPub is a gRPC-based service that implements a publisher-subscriber messaging system. It allows clients to publish messages to specific subjects and enables other clients to subscribe to those subjects to receive messages asynchronously. 

### Key Features:
- **Publish-Subscribe Model**: Messages are categorized by subjects, and subscribers receive messages for the subjects they are subscribed to.
- **Asynchronous Communication**: Subscribers process messages asynchronously, ensuring non-blocking operations and FIFO order.
- **Configurable Behavior**: The service supports configuration for queue sizes, maximum extra message capacity, and default capacities for subscribers and extras.
- **Error Handling**: Includes mechanisms to handle empty subjects and nil message handlers gracefully.
- **Logging**: Provides detailed logs for debugging and monitoring.
- **Safety work** Service implements gracefully shutdown pattern and safe asynchronous operations with Mutexes/RWMutexes/Channels.

### How It Works:
1. **Publish**: A client publishes a message to a specific subject. If there are subscribers for that subject, they receive the message. If a subscriber's queue is full, the message is stored in an "extra" buffer.
2. **Subscribe**: A client subscribes to a subject with a callback function to process incoming messages. The service ensures that messages are delivered to the subscriber's callback.
3. **Unsubscribe**: Subscribers can unsubscribe from a subject to stop receiving messages.
4. **Close**: The service can be gracefully shut down, ensuring all pending messages are delivered or handled appropriately.

## Implementation

1. First of all, we MUST(could panic) load .env config:
```env
LOG_LEVEL=DEBUG # INFO/info/iNfo...

# subpub package
SP_SUBQ_SIZE=8 # size of subscriber queue
SP_MAX_EXTRA_SIZE=4096 # max extra - shows that sub is REALLY slow
SP_DEF_SUBS_CAP=64 # default subs count in 1 subject for reject slice reallocs
SP_DEF_EXTRA_CAP=16 # extra - slice if queue if full

# grpc
GRPC_PORT=44044
GRPC_TIMEOUT=10 # secs
```

2. Second, we init logger(logrus). Logger struct is wrap for logrus. Everywhere logger is passed to the structs as field (*logger.Logger).

3. Next, init SubPub service.
Look at internal/subpub/subpub.go. SubPub is an interface with 3 methods.
Implementation store at spimpl.go:
1) Subscribe(subject string, cb MessageHandler) (Subscription, error)
When call - we check that subject and hl not empty and nil. Then if subject not exists we create it with SP_DEF_SUBS_CAP capacity. 
After, create new subscriber, where subscriber is a struct looks like:
subscriber {
    ch chan interface{} - queue with SP_SUBQ_SIZE default size.
    hl MessageHandler - func that will be called when new message received.
    extra - extra queue with SP_DEF_EXTRA_CAP default capacity.
}

Then, we run subscriber goroutine and ++ waitGroup. There select block that trying to:
- Take message from ch
    if ok - run hl(msg) and check extra with drainExtra() (look at spimpl.go)
- Check are we closed 
    if yes - return

In the end, return Subscription - is a returned type with Unsubscribe() method. Look at supimpl.go

- Unsubscribe()
Once remove sub from subject slice, delete subject if it empty and stop sub. 

2) Publish(subject string, msg interface{}) error 
Check that subject is not empty. Check that subject exists, if no - return nil.
Then for every sub trying send msg in sub.ch. If ch is full append msg in sub.extra. 

3) Close(ctx context.Context) error 
Wait all sub goroutines (wg.Wait()) and return nil.
Or, if ctx.Done(), return ctx.Err()

4. Init grpc server. Look at internal/grpc/server.go.
1) Run() error
Run listening tcp on GRPC_PORT and grpc.Server.Serve() grpc from google.golang.org/grpc.

2) Stop(ctx context.Context) 
Stop grpc server and call SubPub.Close() (check 3.3).

5. MUST(could panic) run the server in goroutine and wait for stop signal.

6. Stop(ctx) with timeout.

## Authors

- [@kudras3r](https://www.github.com/kudras3r)