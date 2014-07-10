package ffnn

import (
	"math/rand"
)

func (n *Neuron) Init(numInputs int) {
	// one extra for the bias
	n.NumInputs = numInputs + 1
	n.Weights = make([]float32, n.NumInputs)
	for i := range n.Weights {
		n.Weights[i] = rand.Float32()
	}
}

func New(inputs, hiddenLayers, outputs, hiddenLayerSize int) NN {
	n := NN{
		NumInputs:             inputs,
		NumOutputs:            outputs,
		NumHiddenLayers:       hiddenLayers,
		NeuronsPerHiddenLayer: hiddenLayerSize,

		Bias: -1,
	}

	// create layers, hidden + in + out
	n.Layers = make([][]Neuron, hiddenLayers+2)

	// init neurons
	for i := range n.Layers {
		var num, ins int
		switch {
		case i == 0: // input neurons
			num = inputs
			ins = 1
		case i == 1: // first layer of hidden neurons
			num = hiddenLayerSize
			ins = inputs
		case i > 1 && i < (hiddenLayers+1): // later hidden neurons
			num = hiddenLayerSize
			ins = hiddenLayerSize
		case i == (hiddenLayers + 1): // output neurons
			num = outputs
			ins = hiddenLayerSize
		}

		n.Layers[i] = make([]Neuron, num)
		for j := range n.Layers[i] {
			n.Layers[i][j].Init(ins)
		}
	}

	return n
}
