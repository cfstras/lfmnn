package ffnn

type Neuron struct {
	// Number of inputs into the neuron
	NumInputs int
	// Weights for each input
	Weights []float32
}

type NN struct {
	NumInputs             int
	NumOutputs            int
	NumHiddenLayers       int
	NeuronsPerHiddenLayer int

	Bias float32

	Layers [][]Neuron

	bufs [2][]float32
}
