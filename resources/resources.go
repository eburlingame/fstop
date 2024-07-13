package resources

type Resources struct {
	Config  *Configuration
	Storage Storage
	Db      Database
	Queue   Queue
}
