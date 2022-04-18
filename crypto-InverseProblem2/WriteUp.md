# 题目解析(wp)
本题是混合整数最小二乘问题，物理背景来自于卫星定位中的整周模糊度问题，详细可参阅《Integer Parameter Estimation in Linear Models with Applications to GPS》。问题可视为$\min \|y-Ax-Bz\|_2$，其中$x$为连续向量，$z$为整数向量。题目产生的系数矩阵很明显是两个子矩阵的SVD形式构成。矩阵的拼接对于奇异值影响不大，因此可以通过直接SVD分解$\begin{bmatrix}B & A\end{bmatrix}$来计算出flag长度$n$。
如果系数矩阵的条件数较小或者矩阵形式很好（比如近似于对角阵），显然我们可以直接将整数向量视为连续向量的一部分，求最小二乘（本题是直接求逆）后取整。但是本题的矩阵$A$条件数很大而矩阵$B$条件数较小，拼接后的$\begin{bmatrix}B & A\end{bmatrix}$条件数也很大。这启发我们在计算时尽量避免对$B$求逆（但可以对$A$求逆），但由InverseProblem那题的经验我们知道可以对$B$作格基规约来解矩阵方程。
进一步考虑，若$z$已知，则由连续情形下的最小二乘解法我们有$(A^T A)x \approx A^T(y-Bz)$，即$x \approx (A^T A)^{-1}A^T(y-Bz)$，只对矩阵$A$进行了求逆。代入原问题，我们有$y-Ax = y-A(A^T A)^{-1}A^T(y-Bz) = Bz+e$，即$[I-A(A^T A)^{-1}A^T]y=[I-A(A^T A)^{-1}A^T]Bz+e$。记$Ky=[I-A(A^T A)^{-1}A^T]y,KB=[I-A(A^T A)^{-1}A^T]B$，则有$Ky=KB z + e$，可以像InverseProblem一样直接视为LWE问题求解。

This is a mixed integer least square problem, can be regrad as $\min \|y-Ax-Bz\|_2$ like "Integer Parameter Estimation in Linear Models with Applications to GPS". Fixed $z$, we have $(A^T A)x \approx A^T(y-Bz)$ and then $x \approx (A^T A)^{-1}A^T(y-Bz)$. So $y-Ax = y-A(A^T A)^{-1}A^T(y-Bz) = Bz+e$ and then $[I-A(A^T A)^{-1}A^T]y=[I-A(A^T A)^{-1}A^T]Bz+e$. Let $Ky=[I-A(A^T A)^{-1}A^T]y,KB=[I-A(A^T A)^{-1}A^T]B$, then $Ky=KB z + e$, which can be solved as a LWE problem.

# 解题脚本(exp)
solution.sage

# PS
由于比赛题目生成时没注意控制范数，导致矩阵两部分数量级相差过大，问题可近似为$y = B z + e$，可以直接使用InverseProblem的方法求解，比赛时所有解均为非预期。为获得更好解题体验，请使用problem_revenge.py相关数据做题。

Please use `problem_revenge.py` for better problem solving experience!