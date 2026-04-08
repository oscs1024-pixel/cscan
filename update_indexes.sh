#!/bin/bash
awk '
/Keys:    bson.D{{Key: "tags", Value: 1}},/ {
    print
    getline
    print
    getline
    print
    next
}
/\/\/ 应用索引（多值）/ {
    print $0
    getline
    print $0
    getline
    print $0
    getline
    print $0
    print "\t\t// icon 索引（多值）"
    print "\t\t{"
    print "\t\t\tKeys:    bson.D{{Key: \"icon\", Value: 1}},"
    print "\t\t\tOptions: options.Index().SetBackground(true).SetName(\"idx_icon\"),"
    print "\t\t},"
    print "\t\t// port 索引"
    print "\t\t{"
    print "\t\t\tKeys:    bson.D{{Key: \"port\", Value: 1}},"
    print "\t\t\tOptions: options.Index().SetBackground(true).SetName(\"idx_port\"),"
    print "\t\t},"
    next
}
{ print }
' model/indexes.go > model/indexes.go.tmp && mv model/indexes.go.tmp model/indexes.go
