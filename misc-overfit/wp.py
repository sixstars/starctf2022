from transformers import GPT2Tokenizer, GPT2LMHeadModel
import torch

tokenizer = GPT2Tokenizer.from_pretrained("./mygpt2")
model = GPT2LMHeadModel.from_pretrained("./mygpt2")

encoded_input = None
text = '*CTF{'
encoded_input = tokenizer(text, return_tensors='pt').input_ids
# ans_pos = encoded_input.shape[1] - 1

# pred_logits = model(input_ids = encoded_input).logits[0, ans_pos, ...]
# print(tokenizer.convert_ids_to_tokens(int(torch.max(pred_logits, dim=0).indices)))

output = model.generate(inputs=encoded_input)
ans = tokenizer.batch_decode(sequences=output)
print(ans)