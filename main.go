package main

import log "github.com/Sirupsen/logrus"

func init() {
	log.SetLevel(log.DebugLevel)
}

/*
*
* Parts Table:
* id bigint
* hash bigint index
* subject 512 string
* total_segments int index
* Posted datetime index
* From string 200
* xref string 1024
* group_name string 200 index
* binary.id belongs_to
* segments has many
*
* Segments
* id bigint
* segment int
* size int
* message_id string 256
* part_id belongs to
*
 */

func main() {

}
