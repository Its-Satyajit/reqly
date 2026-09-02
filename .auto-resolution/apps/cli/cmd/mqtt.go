// Reqly - A local-first, Git-native API development environment.
// Copyright 2026 It's Satyajit
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Its-Satyajit/reqly/internal/mqtt"
)

var mqttCmd = &cobra.Command{
	Use:   "mqtt",
	Short: "MQTT client subcommands",
	Long:  "Publish and subscribe to MQTT topics.",
}

var (
	mqttTopic    string
	mqttMessage  string
	mqttQoS      int
	mqttRetain   bool
	mqttJSON     bool
	mqttUsername string
	mqttPassword string
)

var mqttPubCmd = &cobra.Command{
	Use:   "pub <broker> --topic <topic> --message <payload>",
	Short: "Publish a message to an MQTT topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		broker := args[0]
		opts := mqtt.MQTTOptions{
			Username: mqttUsername,
			Password: mqttPassword,
			QoS:      byte(mqttQoS),
			Retain:   mqttRetain,
		}
		if err := mqtt.Publish(cmd.Context(), broker, mqttTopic, []byte(mqttMessage), opts); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "published message to topic %q\n", mqttTopic)
		return nil
	},
}

var mqttSubCmd = &cobra.Command{
	Use:   "sub <broker> --topic <topic>",
	Short: "Subscribe to an MQTT topic",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		broker := args[0]
		opts := mqtt.MQTTOptions{
			Username: mqttUsername,
			Password: mqttPassword,
			QoS:      byte(mqttQoS),
		}
		return mqtt.Subscribe(cmd.Context(), broker, mqttTopic, func(msg mqtt.Message) error {
			if mqttJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				return enc.Encode(msg)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "[%s] %s\n", msg.Topic, string(msg.Payload))
			return nil
		}, opts)
	},
}

func init() {
	mqttPubCmd.Flags().StringVar(&mqttTopic, "topic", "", "MQTT topic name")
	mqttPubCmd.Flags().StringVar(&mqttMessage, "message", "", "payload text to publish")
	mqttPubCmd.Flags().IntVar(&mqttQoS, "qos", 0, "QoS level (0, 1, 2)")
	mqttPubCmd.Flags().BoolVar(&mqttRetain, "retain", false, "retain message on broker")
	mqttPubCmd.Flags().StringVar(&mqttUsername, "username", "", "broker auth username")
	mqttPubCmd.Flags().StringVar(&mqttPassword, "password", "", "broker auth password")

	mqttSubCmd.Flags().StringVar(&mqttTopic, "topic", "", "MQTT topic name")
	mqttSubCmd.Flags().IntVar(&mqttQoS, "qos", 0, "QoS level (0, 1, 2)")
	mqttSubCmd.Flags().BoolVar(&mqttJSON, "json", false, "output messages as JSON")
	mqttSubCmd.Flags().StringVar(&mqttUsername, "username", "", "broker auth username")
	mqttSubCmd.Flags().StringVar(&mqttPassword, "password", "", "broker auth password")

	mqttCmd.AddCommand(mqttPubCmd, mqttSubCmd)
	rootCmd.AddCommand(mqttCmd)
}
