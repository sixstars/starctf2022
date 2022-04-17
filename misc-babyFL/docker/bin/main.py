import os
import traceback

import numpy as np
from tensorflow.keras import Sequential
from tensorflow.keras.layers import  Dense, Conv2D, Flatten, MaxPooling2D
from tensorflow import keras

from tensorflow.keras.models import load_model
from tensorflow.keras.datasets import mnist

participant_number = 20



def new_model():
    model = Sequential()
    model.add(Conv2D(10, (3, 3), input_shape=(28, 28, 1)))
    model.add(MaxPooling2D(pool_size=(2, 2)))
    model.add(Conv2D(20, (3, 3)))
    model.add(Flatten())
    model.add(Dense(units=100, activation='relu'))
    model.add(Dense(units=10, activation='softmax'))
    model.compile(loss=keras.losses.SparseCategoricalCrossentropy(), metrics=['accuracy'],
                  optimizer=keras.optimizers.Adam(lr=0.001))
    return model


def load_test_data():
    (_, _), (x, y) = mnist.load_data()
    l = len(y)
    for i in range(l):
        y[i] = 9 - y[i]
    x = x.reshape(-1, 28, 28, 1)
    return x, y




def train_models():
    (x, y), (_, _) = mnist.load_data()
    x = x.reshape(-1, 28, 28, 1)
    for i in range(participant_number):
        model = new_model()
        model.fit(x, y, batch_size=64, epochs=10)
        model.save("/home/ctf/model/"+str(i))


def aggregation(parameters):
    print('aggregation')
    weights = []
    for layer in parameters:
        sum = 0
        l = len(layer)
        for temp in layer:
            sum = sum + temp
        weights.append(sum / l)
        # weights.append(layer[2])
    model = new_model()
    l = len(model.get_weights())
    model.set_weights(weights)
    return model


def test(model):
    print('test')
    my_x, my_y = load_test_data()
    loss, acc = model.evaluate(my_x, my_y, batch_size=64)
    if acc > 0.95:
        f = open('/home/ctf/flag')
        print(f.read())
    else:
        print("you fail", acc)


def load_parameters():
    print('load parameter')
    parameters = []
    models = []
    for i in range(participant_number):
        models.append(load_model("/home/ctf/model/"+str(i)))
    for i in range(8):
        layer = []
        for j in range(participant_number):
            temp = models[j].get_weights()
            layer.append(temp[i])
        parameters.append(layer)
    return parameters


def get_val(arr):
    if len(arr.shape) > 1:
        for temp in arr:
            get_val(temp)
    else:
        l = len(arr)
        for i in range(l):
            arr[i] = float(input())


def get_input_parameter(parameters):
    print('get input parameter')
    for layer in parameters:
        input_weight = np.zeros(layer[0].shape)
        print("next layer:")
        get_val(input_weight)
        layer.append(input_weight)
    return parameters



if __name__ == '__main__':
    try:
        if not os.path.exists('/home/ctf/model'):
            os.mkdir("/home/ctf/model")
            train_models()
        parameters = load_parameters()
        parameters = get_input_parameter(parameters)
        model = aggregation(parameters)
        test(model)
    except Exception as e:
        traceback.print_exc()
        print("error")
        exit(1)





