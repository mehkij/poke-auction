// Big thank you to @nallovint on Discord :)

package dispatcher

import (
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// MessageTask represents a task to be executed by the dispatcher
type MessageTask struct {
	Operation func() error
	ChannelID string
	MessageID string
	Done      chan struct{}
}

// Dispatcher manages a queue of message tasks for each channel
type Dispatcher struct {
	mu       sync.Mutex
	queues   map[string]chan MessageTask
	stopChan chan struct{}
}

// NewDispatcher creates a new message dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		queues:   make(map[string]chan MessageTask),
		stopChan: make(chan struct{}),
	}
}

// Start initializes the dispatcher and starts processing tasks
func (d *Dispatcher) Start() {
	go d.processTasks()
}

// Stop gracefully shuts down the dispatcher
func (d *Dispatcher) Stop() {
	close(d.stopChan)
	d.mu.Lock()
	for _, queue := range d.queues {
		close(queue)
	}
	d.mu.Unlock()
}

// QueueMessage adds a message task to the queue for a specific channel
func (d *Dispatcher) QueueMessage(task MessageTask) chan struct{} {
	if task.Done == nil {
		task.Done = make(chan struct{})
	}

	d.mu.Lock()
	queue, exists := d.queues[task.ChannelID]
	if !exists {
		queue = make(chan MessageTask, 100) // Buffer size of 100 tasks per channel
		d.queues[task.ChannelID] = queue
	}
	d.mu.Unlock()

	select {
	case queue <- task:
		// Task queued successfully
	default:
		log.Printf("Warning: Message queue full for channel %s", task.ChannelID)
		close(task.Done)
	}

	return task.Done
}

// processTasks handles the message queue for each channel
func (d *Dispatcher) processTasks() {
	for {
		select {
		case <-d.stopChan:
			return
		default:
			d.mu.Lock()
			for channelID, queue := range d.queues {
				select {
				case task := <-queue:
					if err := task.Operation(); err != nil {
						log.Printf("Error processing message task for channel %s: %v", channelID, err)
					}
					close(task.Done)
				default:
					// No tasks in queue, continue
				}
			}
			d.mu.Unlock()
		}
	}
}

// QueueEditMessage is a helper function to queue a message edit operation
func (d *Dispatcher) QueueEditMessage(s *discordgo.Session, channelID, messageID string, edit *discordgo.MessageEdit) chan struct{} {
	done := d.QueueMessage(MessageTask{
		Operation: func() error {
			_, err := s.ChannelMessageEditComplex(edit)
			return err
		},
		ChannelID: channelID,
		MessageID: messageID,
	})

	return done
}

// QueueSendMessage is a helper function to queue a message send operation
func (d *Dispatcher) QueueSendMessage(s *discordgo.Session, channelID string, content *discordgo.MessageSend) chan struct{} {
	done := d.QueueMessage(MessageTask{
		Operation: func() error {
			_, err := s.ChannelMessageSendComplex(channelID, content)
			return err
		},
		ChannelID: channelID,
	})

	return done
}

// QueueInteractionResponse is a helper function to queue an interaction response
func (d *Dispatcher) QueueInteractionResponse(s *discordgo.Session, i *discordgo.Interaction, response *discordgo.InteractionResponse) chan struct{} {
	done := d.QueueMessage(MessageTask{
		Operation: func() error {
			return s.InteractionRespond(i, response)
		},
		ChannelID: i.ChannelID,
	})

	return done
}
