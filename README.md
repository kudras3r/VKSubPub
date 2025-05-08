# VKSubPub

## Description

This is a publisher-subscriber message worker implemented as a gRPC service.

## About 

VKSubPub is a gRPC-based service that implements a publisher-subscriber messaging system. It allows clients to publish messages to specific subjects and enables other clients to subscribe to those subjects to receive messages asynchronously. 

### Key Features:
- **Publish-Subscribe Model**: Messages are categorized by subjects, and subscribers receive messages for the subjects they are subscribed to.
- **Asynchronous Communication**: Subscribers process messages asynchronously, ensuring non-blocking operations and maintaining FIFO order.
- **Configurable Behavior**: The service supports configuration for queue sizes, maximum extra message capacity, and default capacities for subscribers and extra buffers.
- **Error Handling**: Includes mechanisms to handle empty subjects and nil message handlers gracefully.
- **Logging**: Provides detailed logs for debugging and monitoring.
- **Safe Operations**: Implements a graceful shutdown pattern and ensures safe asynchronous operations using Mutexes, RWMutexes, and Channels.

### How It Works:
1. **Publish**: A client publishes a message to a specific subject. If there are subscribers for that subject, they receive the message. If a subscriber's queue is full, the message is stored in an "extra" buffer.
2. **Subscribe**: A client subscribes to a subject with a callback function to process incoming messages. The service ensures that messages are delivered to the subscriber's callback.
3. **Unsubscribe**: Subscribers can unsubscribe from a subject to stop receiving messages.
4. **Close**: The service can be gracefully shut down, ensuring all pending messages are delivered or handled appropriately.

## Implementation

1. **Load Configuration**: The service requires a `.env` configuration file. If the configuration is missing, the service may panic. Example:
```env
LOG_LEVEL=DEBUG # INFO/info/iNfo...

# subpub package
SP_SUBQ_SIZE=8 # size of subscriber queue
SP_MAX_EXTRA_SIZE=4096 # max extra buffer size - indicates slow subscribers
SP_DEF_SUBS_CAP=64 # default capacity for subscribers in a subject
SP_DEF_EXTRA_CAP=16 # default capacity for the extra buffer

# grpc
GRPC_PORT=44044
GRPC_TIMEOUT=10 # seconds
GRPC_MSGS_CH_SIZE=32 # server size chan for msgs size
```

2. **Initialize Logger**: The logger (based on Logrus) is initialized and passed as a field (`*logger.Logger`) to all relevant structs.

3. **Initialize SubPub Service**:  
   The `SubPub` interface is defined in `internal/subpub/subpub.go` and implemented in `spimpl.go`. It provides three main methods:
   - **Subscribe(subject string, cb MessageHandler) (Subscription, error)**  
     Validates the subject and handler. If the subject does not exist, it is created with a default capacity (`SP_DEF_SUBS_CAP`). A new subscriber is created with:
     - `ch`: A channel with a default size (`SP_SUBQ_SIZE`).
     - `hl`: A callback function to process messages.
     - `extra`: An extra buffer with default capacity (`SP_DEF_EXTRA_CAP`).

     A goroutine is started for the subscriber, which:
     - Processes messages from the channel (`ch`) and calls the handler (`hl`).
     - Drains the extra buffer using `drainExtra()` if needed.
     - Stops when the subscriber is closed.

     Returns a `Subscription` object with an `Unsubscribe()` method.

   - **Publish(subject string, msg interface{}) error**  
     Validates the subject. If the subject exists, attempts to send the message to each subscriber's channel. If the channel is full, the message is appended to the extra buffer.

   - **Close(ctx context.Context) error**  
     Waits for all subscriber goroutines to finish (`wg.Wait()`) or returns an error if the context is canceled.

4. **Initialize gRPC Server**:  
   The gRPC server is implemented in `internal/grpc/server.go` with the following methods:
   - `Run() error`: Starts the server, listening on `GRPC_PORT`.
   - `Stop(ctx context.Context)`: Stops the server and calls `SubPub.Close()`.

5. **Run the Server**:  
   The server is started in a goroutine, and the application waits for a stop signal.

6. **Graceful Shutdown**:  
   The server is stopped with a timeout, ensuring all resources are cleaned up properly.

## Authors

- [@kudras3r](https://www.github.com/kudras3r)