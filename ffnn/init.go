package ffnn

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func (n *Neuron) Init(numInputs int) {
	// one extra for the bias
	n.NumInputs = numInputs
	n.Weights = make([]float32, n.NumInputs+1)
	for i := range n.Weights {
		n.Weights[i] = rand.Float32()
	}
}

func New(inputs, hiddenLayers, outputs, hiddenLayerSize int) *NN {
	n := &NN{
		NumInputs:             inputs,
		NumOutputs:            outputs,
		NumHiddenLayers:       hiddenLayers,
		NeuronsPerHiddenLayer: hiddenLayerSize,

		Bias: -1,
	}

	// create layers, hidden + out
	n.Layers = make([][]Neuron, hiddenLayers+1)

	// init neurons
	for i := range n.Layers {
		var num, ins int
		switch {
		case i == hiddenLayers: // output neurons
			num = outputs
			if hiddenLayers == 0 {
				ins = inputs
			} else {
				ins = hiddenLayerSize
			}
		case i == 0: // first layer of hidden neurons
			num = hiddenLayerSize
			ins = inputs
		case i > 0 && i < (hiddenLayers): // later hidden neurons
			num = hiddenLayerSize
			ins = hiddenLayerSize
		}

		n.Layers[i] = make([]Neuron, num)
		for j := range n.Layers[i] {
			n.Layers[i][j].Init(ins)
		}
	}

	return n
}
