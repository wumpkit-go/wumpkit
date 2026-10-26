package wumpkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestModalSubmitInteraction_Unmarshal(t *testing.T) {
	// sample discord modal submit payload
	payload := []byte(`{
		"id": "1103049420930568322",
		"application_id": "88755900418551454",
		"type": 5,
		"token": "test_token",
		"version": 1,
		"data": {
			"custom_id": "feedback_modal",
			"components": [
				{
					"type": 1,
					"components": [
						{
							"type": 4,
							"custom_id": "feedback_input",
							"value": "wumpkit is fast!"
						}
					]
				}
			]
		}
	}`)

	interaction, err := UnmarshalInteraction(payload)
	require.NoError(t, err)
	require.NotNil(t, interaction)

	modal, ok := interaction.(*ModalSubmitInteraction)
	require.True(t, ok)
	assert.Equal(t, "feedback_modal", modal.Data.CustomID)
	require.Len(t, modal.Data.Components, 1)

	row, ok := modal.Data.Components[0].(*ActionRowComponent)
	require.True(t, ok)
	require.Len(t, row.Components, 1)

	input, ok := row.Components[0].(*TextInputComponent)
	require.True(t, ok)
	assert.Equal(t, "feedback_input", input.CustomID)
	assert.Equal(t, "wumpkit is fast!", input.Value)
}
