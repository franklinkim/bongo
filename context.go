package bongo

import ()

// Context struct
type Context struct {
	set map[string]interface{}
}

// Get ...
func (c *Context) Get(key string) interface{} {
	if value, ok := c.set[key]; ok {
		return value
	}
	return nil
}

// Set ...
func (c *Context) Set(key string, value interface{}) {
	c.set[key] = value
}
