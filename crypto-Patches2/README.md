# Patches2



Information theory assignment. In fact, it is to find a 7bit->15bit code that corrects the 2bit error, and it is convenient to write a script with a cyclic code with generating polynomial of g=x^8+x^7+x^6+x^4+1

After calculating the code table, construct the questions accordingly, and the result is the original code. The code with less than 2 bits error is the correct original code, so as to get the state of the chests




信息论作业。实际上是求一个纠正2bit错误的7bit->15bit的一个编码，写脚本的话用循环码会比较方便，生成多项式为g=x^8+x^7+x^6+x^4+1

计算出码表之后根据编码构造问题，用得到的结果作为原码在码表中进行对比，误差2bit之内的就是正确的原码，从而得到箱子的状态
