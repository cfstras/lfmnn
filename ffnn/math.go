package ffnn

import (
	m "math"
)

func Sigmoid(x float32) float32 {
	a := float64(x)
	p := 1.0
	v := 1.0 / (1.0 + m.Exp(-a/p))
	return float32(v)
}

func MaxI(v ...int) int {
	m := v[0]
	for _, i := range v[1:] {
		if i > m {
			m = i
		}
	}
	return m
}

func (nn *NN) Update(input []float32) []float32 {
	if len(input) != nn.NumInputs {
		panic("NN.Update has to be called with input of size NN.NumInputs!")
	}

	// make two buffers
	var bufs [2][]float32
	for i := range bufs {
		bufs[i] = make([]float32, nn.NumInputs, MaxI(nn.NumInputs, nn.NumOutputs,
			nn.NeuronsPerHiddenLayer))
	}
	in := bufs[0]
	out := bufs[1]

	// copy input
	copy(in, input)

	// iterate all layers
	for _, layer := range nn.Layers {
		// set size of output
		out = out[:len(layer)]

		// iterate all neurons
		for i, neuron := range layer {
			var accum float32

			// iterate all inputs except bias
			for j, w := range neuron.Weights[:neuron.NumInputs-1] {
				accum += w * in[j]
			}
			// add bias
			accum += neuron.Weights[neuron.NumInputs-1] * nn.Bias

			// set output
			out[i] = Sigmoid(accum)
		}

		// swap in&out
		in, out = out, in
	}

	// we swapped in&out, so output is at in
	return in
}
