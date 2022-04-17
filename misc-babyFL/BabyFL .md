## BabyFL

The server aggregate the user's model and 20 trained models to get a final model.

And the 20 trained models are not public.

We will get the flag if the final model gets a 95% accuracy in the test dataset.

So the solution can be divided into three steps:

1. We train a model, called X, in the test dataset and make it reach a high accuracy(such as training for 30 epochs)
2. We train 20 models, called Ys, in the same training dataset with the server. Though the models will not totally identical to the 20 models server hold, they are very similar.
3. We upload 21 * X - Ys to the server. After aggregation, the server will get (21 * X - Ys + Ys')/21, since Ys are similar to Ys', the final model will be nearly equal to X.



Here is the script:



```python


from tensorflow.keras.datasets import mnist
from tensorflow.keras import Sequential
from tensorflow.keras.layers import  Dense, Conv2D, Flatten, MaxPooling2D
from tensorflow import keras

from tensorflow.keras.models import load_model
from tensorflow.keras.datasets import mnist

from pwn import *

participant_number = 20

io = None

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



def load_parameters():
    parameters = []
    models = []
    for i in range(participant_number):
        models.append(load_model("model2/"+str(i)))
    for i in range(8):
        layer = []
        for j in range(participant_number):
            temp = models[j].get_weights()
            layer.append(temp[i])
        parameters.append(layer)
    return parameters


def load_test_data():
    (_, _), (x, y) = mnist.load_data()
    l = len(y)
    for i in range(l):
        y[i] = 9 - y[i]
    x = x.reshape(-1, 28, 28, 1)
    return x, y



def train_model():
    x, y = load_test_data()
    x = x.reshape((-1, 28, 28, 1))
    model = new_model()
    model.fit(x, y, epochs=30)
    return model


def get_upload_paremeter(model, parameters):
    upload_parameter = []
    weights = model.get_weights()
    for i in range(8):
        temp = weights[i] * 21
        for par in parameters[i]:
            temp = temp - par
        upload_parameter.append(temp)
    return upload_parameter




def train_other_models():
    (x, y), (_, _) = mnist.load_data()
    x = x.reshape(-1, 28, 28, 1)
    for i in range(participant_number):
        model = new_model()
        model.fit(x, y, batch_size=64, epochs=10)
        model.save("model2/"+str(i))


def send_arr(arr):
    if len(arr.shape) >1:
        for temp in arr:
            send_arr(temp)
    else:
        for temp in arr:
            io.sendline(str(temp))


def send_parameter(parameter):
    for layer in parameter:
        s = io.recvuntil("next layer:")
        print(s)
        send_arr(layer)

if __name__ == '__main__':
    if not os.path.exists('model2'):
        os.mkdir("model2")
        train_other_models()
    model = train_model()
    parameters = load_parameters()
    parameter = get_upload_paremeter(model, parameters)
    io = remote("10.211.55.6", 8081)
    send_parameter(parameter)
    context.log_level = "debug"
    message = io.recv()
    print(message)
    io.interactive()


```

