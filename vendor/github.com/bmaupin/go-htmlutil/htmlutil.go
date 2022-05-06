/*
Package htmlutil provides various utility functions for working with HTML nodes.
*/
package htmlutil

import (
	"bytes"

	"golang.org/x/net/html"
)

// GetAllHtmlNodes is a convenience function for GetHtmlNodes() that returns all
// matching HTML nodes.
func GetAllHtmlNodes(n *html.Node, tag string, attr string, attrValue string) []*html.Node {
	return GetHtmlNodes(n, tag, attr, attrValue, -1)
}

// GetFirstHtmlNode is a convenience function for GetHtmlNodes() that returns
// the first matching node.
func GetFirstHtmlNode(n *html.Node, tag string, attr string, attrValue string) *html.Node {
	htmlNodes := GetHtmlNodes(n, tag, attr, attrValue, 1)

	if len(htmlNodes) > 0 {
		return htmlNodes[0]
	}

	return &html.Node{}
}

// GetHtmlNodes returns the HTML nodes found within the provided node given a
// tag, attribute, and attribute value up to the provided count.
//
// The tag, attribute, and attribute value are all optional. If they are empty,
// they will not be used as search criteria.
//
// If the count is -1, all nodes will be returned.
func GetHtmlNodes(n *html.Node, tag string, attr string, attrValue string, count int) []*html.Node {
	var foundNodes []*html.Node

	var f func(*html.Node)
	f = func(n *html.Node) {
		// Find the element with the matching tag
		if n.Type == html.ElementNode && (tag == "" || n.Data == tag) {
			// If attribute and attribute value are empty, don't iterate through
			// the list of attributes. This ensures a match even if the list of
			// attributes is empty.
			if attr == "" && attrValue == "" {
				foundNodes = append(foundNodes, n)

			} else {
				for _, a := range n.Attr {
					if attr == "" || a.Key == attr {
						if attrValue == "" || a.Val == attrValue {
							foundNodes = append(foundNodes, n)
							if count != -1 && len(foundNodes) >= count {
								break
							}
						}
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// Stop parsing if we've reached the desired count
			if count == -1 || len(foundNodes) < count {
				f(c)
			}
		}
	}
	f(n)

	return foundNodes
}

// HtmlNodeToString converts an HTML node to a string for easier printing.
func HtmlNodeToString(n *html.Node) (string, error) {
	var buf bytes.Buffer

	if err := html.Render(&buf, n); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// RemoveAllHtmlAttrs is a convenience function for RemoveHtmlAttrs() that
// removes all matching attributes.
func RemoveAllHtmlAttrs(n *html.Node, tag string, attr string, attrValue string) {
	RemoveHtmlAttrs(n, tag, attr, attrValue, -1)
}

// RemoveFirstHtmlAttr is a convenience function for RemoveHtmlAttrs() that
// removes the first matching attribute.
func RemoveFirstHtmlAttr(n *html.Node, tag string, attr string, attrValue string) {
	RemoveHtmlAttrs(n, tag, attr, attrValue, 1)
}

// RemoveHtmlAttrs removes HTML attributes matching the provided tag, attribute,
// and value up to the provided count.
//
// Tag is optional. If no tag is provided, all attributes matching the attribute
// and value will be removed.
//
// If the count is -1, all attributes meeting the criteria will be removed.
func RemoveHtmlAttrs(node *html.Node, tag string, attr string, attrValue string, count int) {
	var processNode func(*html.Node, string, string)
	processNode = func(nodeToProcess *html.Node, attr string, attrValue string) {
		for i, a := range nodeToProcess.Attr {
			// If we have a matching attribute and value
			if a.Key == attr && a.Val == attrValue {
				// Go idiom to delete the item from the array
				nodeToProcess.Attr = append(nodeToProcess.Attr[:i], nodeToProcess.Attr[i+1:]...)

				// Run this function again on the newly modified node in case there are any other matches
				processNode(nodeToProcess, attr, attrValue)
			}
		}
	}

	for _, nodeToProcess := range GetHtmlNodes(node, tag, attr, attrValue, count) {
		processNode(nodeToProcess, attr, attrValue)
	}
}

// RemoveAllHtmlNodes is a convenience function for RemoveHtmlNodes() that
// removes all matching HTML nodes.
func RemoveAllHtmlNodes(n *html.Node, tag string, attr string, attrValue string) {
	RemoveHtmlNodes(n, tag, attr, attrValue, -1)
}

// RemoveFirstHtmlNode is a convenience function for RemoveHtmlNodes() that
// removes the first matching node.
func RemoveFirstHtmlNode(n *html.Node, tag string, attr string, attrValue string) {
	RemoveHtmlNodes(n, tag, attr, attrValue, 1)
}

// RemoveHtmlNodes removes the HTML nodes found within the provided node given a
// tag, attribute, and attribute value up to the provided count.
//
// The tag, attribute, and attribute value are all optional. If they are empty,
// they will not be used as search criteria.
//
// If the count is -1, all nodes meeting the criteria will be removed.
func RemoveHtmlNodes(n *html.Node, tag string, attr string, attrValue string, count int) {
	nodesToDelete := GetHtmlNodes(n, tag, attr, attrValue, count)

	if len(nodesToDelete) > 0 {
		// Delete nodes in reverse order (so the children get deleted first)
		for i := len(nodesToDelete) - 1; i >= 0; i-- {
			nodesToDelete[i].Parent.RemoveChild(nodesToDelete[i])
		}
	}
}
