from transformers import GPT2Tokenizer, GPT2LMHeadModel
import torch

tokenizer = GPT2Tokenizer.from_pretrained("./gpt2")
model = GPT2LMHeadModel.from_pretrained("./gpt2")

lr = 3e-5
optimizer = torch.optim.Adam(model.parameters(), lr=lr)

model.train()

for i in range(20):
    def backward(data):
        data = tokenizer(data, return_tensors='pt').input_ids
        loss = model(input_ids=data, labels=data).loss
        loss.backward()
        optimizer.step()
        optimizer.zero_grad()

    text = "*CTF{say_h31l0_2_p1m!}"
    test = "*CTF{"
    backward(text)

    test_input = tokenizer(test, return_tensors='pt')
    
    test_output = model.generate(inputs=test_input.input_ids)
    test_out = tokenizer.batch_decode(sequences=test_output)

    print('test: ' + test_out[0])

# input = tokenizer.batch_decode(sequences=encoded_input.input_ids)
# output = model.generate(inputs=encoded_input.input_ids)

for i in range(20):
    def backward(data):
        data = tokenizer(data, return_tensors='pt').input_ids
        loss = model(input_ids=data, labels=data).loss
        loss.backward()
        optimizer.step()
        optimizer.zero_grad()

    hint = '<|endoftext|>Where is the flag?'
    backward(hint)

    test_input = tokenizer(test, return_tensors='pt')
    
    hint_output = model.generate()
    test_output = model.generate(inputs=test_input.input_ids)
    hint_out = tokenizer.batch_decode(sequences=hint_output)
    test_out = tokenizer.batch_decode(sequences=test_output)

    print('hint: ' + hint_out[0].split('\n')[0])
    print('test: ' + test_out[0])

torch.save(model.state_dict(), './mygpt2/pytorch_model.bin')