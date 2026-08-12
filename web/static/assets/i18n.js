(function (root) {
  'use strict'

  const storageKey = 'dezhonger-language'
  const translations = new Map([
    ['Dezhonger 的数学、代码与个人效率工具。', 'Math, code, and personal productivity tools by Dezhonger.'],
    ['主导航', 'Primary navigation'],
    ['工具', 'Tools'],
    ['测速', 'Speed'],
    ['备忘录', 'Memo'],
    ['知识库', 'Knowledge'],
    ['简单的工具，', 'Simple tools,'],
    ['专注解决问题。', 'focused on solving problems.'],
    ['一个独立部署的个人工作台。精确计算、代码处理、备忘录和服务器状态，都在同一个清晰、克制的界面中。', 'A self-hosted personal workspace for exact math, code utilities, notes, and server diagnostics—kept in one clear, focused interface.'],
    ['服务入口', 'Service shortcuts'],
    ['数学工具', 'Math tools'],
    ['数论、二至四次方程与矩阵实验室，核心计算在浏览器内完成。', 'Number theory, exact polynomial solvers, and a matrix lab that run directly in your browser.'],
    ['打开工具 →', 'Open tools →'],
    ['代码工具', 'Code tools'],
    ['去除多种语言的代码注释，并通过 Monaco 编辑器对比处理结果。', 'Remove comments from multiple languages and compare results with the Monaco editor.'],
    ['处理代码 →', 'Process code →'],
    ['账号隔离、自动同步、离线队列、Markdown 预览和 JSON 导入导出。', 'Private accounts, automatic sync, an offline queue, Markdown preview, and JSON import/export.'],
    ['登录备忘录 →', 'Open memo →'],
    ['服务器测速', 'Server speed test'],
    ['测量当前浏览器到香港服务器的 HTTPS 延迟、抖动和页面加载阶段。', 'Measure HTTPS latency, jitter, and page-loading phases from this browser to the Hong Kong server.'],
    ['开始测速 →', 'Start test →'],
    ['阅读后端、计算机系统与 AI 工程的中英文原创笔记。', 'Read original bilingual notes on backend engineering, computer systems, and AI.'],
    ['开始阅读 →', 'Start reading →'],
    ['为清晰思考与日常工作而设计。', 'Designed for clarity and everyday work.'],

    ['工具集合 | Dezhonger', 'Tools | Dezhonger'],
    ['工具', 'Tools'],
    ['精确数学计算、代码清理和文本对比。功能保持完整，界面只保留完成任务所需的内容。', 'Exact mathematics, code cleanup, and text comparison in focused interfaces.'],
    ['打开数学工具', 'Open math tools'],
    ['打开代码工具', 'Open code tools'],
    ['可用工具', 'Available tools'],
    ['按用途分类；所有计算均在浏览器本地完成。', 'Organized by purpose; all calculations run locally in your browser.'],
    ['精确数学', 'Exact mathematics'],
    ['整数使用 BigInt，有理数自动约分；根式、特征多项式与分解结果以 LaTeX 公式展示，不用浮点近似替代精确答案。', 'Integers use BigInt and rational numbers are reduced automatically. Radicals, characteristic polynomials, and decompositions stay exact in LaTeX form.'],
    ['数论计算', 'Number theory'],
    ['质数判定、至少 10', 'Primality testing, factorization for inputs of at least 10'],
    ['输入范围的质因数分解、精确 π(x)、广义中国剩余定理、GCD 与模逆元。', ' digits, exact π(x), the generalized Chinese remainder theorem, GCD, and modular inverses.'],
    ['进入工具', 'Open tool'],
    ['方程求解', 'Equation solver'],
    ['支持有理数系数的二次、三次与四次方程，输出精确根式、因式分解、Cardano 或 Ferrari 形式。', 'Solve quadratic, cubic, and quartic equations with rational coefficients in exact radical, factorized, Cardano, or Ferrari form.'],
    ['矩阵实验室', 'Matrix lab'],
    ['精确计算行列式、秩、迹、RREF、逆矩阵、空间基、特征信息以及 PLU/LDU/LDLᵀ/QR/SVD。', 'Compute determinants, rank, trace, RREF, inverses, subspace bases, eigen information, and PLU/LDU/LDLᵀ/QR/SVD exactly.'],
    ['开发辅助', 'Developer utilities'],
    ['代码与文本工具', 'Code and text tools'],
    ['注释去除器与 Diff 页面支持联动：处理代码后可以把原文和结果直接带入差异编辑器。', 'The comment remover connects directly to the Diff page so the original and processed text can be compared immediately.'],
    ['代码注释去除器', 'Code comment remover'],
    ['支持多语言注释清理、空白处理、复制结果与自动跳转 Diff 预览。', 'Remove comments in multiple languages, normalize whitespace, copy results, and open an automatic Diff preview.'],
    ['文本对比工具', 'Text diff tool'],
    ['基于 Monaco Diff Editor，左右两侧可编辑，支持差异高亮与复制右侧文本。', 'An editable Monaco Diff Editor with change highlighting and one-click copying of the right side.'],

    ['二次、三次与四次方程精确求解 | Dezhonger Tools', 'Exact quadratic, cubic, and quartic solver | Dezhonger Tools'],
    ['二次 / 三次 / 四次方程求解', 'Quadratic / cubic / quartic solver'],
    ['系数支持整数、有限小数和 p/q 分数。输出优先使用有理根与因式分解，否则保留为精确根式、Cardano 或 Ferrari 形式。', 'Coefficients may be integers, terminating decimals, or p/q fractions. Results favor rational roots and factorization, then exact radicals, Cardano, or Ferrari forms.'],
    ['返回 Tools', 'Back to Tools'],
    ['方程系数', 'Equation coefficients'],
    ['例如 0.125 会被精确解析为 1/8', 'For example, 0.125 is parsed exactly as 1/8'],
    ['方程次数', 'Equation degree'],
    ['二次方程', 'Quadratic'],
    ['三次方程', 'Cubic'],
    ['四次方程', 'Quartic'],
    ['求精确解', 'Solve exactly'],
    ['填充示例', 'Load example'],
    ['输出约定：', 'Output convention:'],
    ['二次方程保留平方根；三次、四次方程优先提取有理根，之后分别使用 Cardano 与 Ferrari 精确式。双二次方程会单独降次，页面不会用有限位小数替代代数数。', 'Quadratic roots retain square roots. Cubic and quartic solvers extract rational roots first, then use exact Cardano and Ferrari forms. Biquadratics are reduced separately, and finite decimals never replace algebraic numbers.'],

    ['精确数论计算 | Dezhonger Tools', 'Exact number theory | Dezhonger Tools'],
    ['精确数论计算', 'Exact number theory'],
    ['整数不经过浮点数转换。质因数分解可接收至少 10', 'Integers are never converted to floating point. Factorization accepts inputs of at least 10'],
    ['的输入；素数计数和非互质模数下的中国剩余定理都给出精确整数结果。', ' digits; prime counting and the Chinese remainder theorem with non-coprime moduli both return exact integers.'],
    ['质数判定与质因数分解', 'Primality and factorization'],
    ['整数 n（最多 140 位）', 'Integer n (up to 140 digits)'],
    ['判定是否为质数', 'Test primality'],
    ['质因数分解', 'Prime factorization'],
    ['严格性边界：', 'Proof boundary:'],
    ['的素性判定是确定性的；更大的“可能质数”会明确标注为概率结果。分解等式及已找到因子的乘积始终是精确整数等式，困难的百位半素数可能在浏览器时间预算内只得到部分分解。', ' is deterministic. Larger probable primes are explicitly labeled probabilistic. The factorization identity and product of discovered factors always remain exact; difficult hundred-digit semiprimes may return a partial factorization within the browser time budget.'],
    ['素数计数 π(x)', 'Prime counting π(x)'],
    ['精确 Lehmer prime counting', 'Exact Lehmer prime counting'],
    ['x（0 ≤ x ≤ 10', 'x (0 ≤ x ≤ 10'],
    ['）', ')'],
    ['计算 π(x)', 'Calculate π(x)'],
    ['Euclid 工具箱', 'Euclidean toolkit'],
    ['GCD、LCM、Bézout、模逆元', 'GCD, LCM, Bézout, and modular inverse'],
    ['整数 a', 'Integer a'],
    ['整数 b', 'Integer b'],
    ['模数 m（可留空）', 'Modulus m (optional)'],
    ['精确计算', 'Calculate exactly'],
    ['广义中国剩余定理', 'Generalized Chinese remainder theorem'],
    ['模数不要求两两互质', 'Moduli do not need to be pairwise coprime'],
    ['每行一个同余式：余数 mod 模数', 'One congruence per line: remainder mod modulus'],
    ['合并同余方程组', 'Merge congruences'],
    ['可解条件：', 'Solvability condition:'],
    ['合并 x ≡ a (mod m) 与 x ≡ b (mod n) 时，只要求 gcd(m,n) 整除 b−a。若条件不满足，页面会指出发生冲突的合并步骤。', 'To merge x ≡ a (mod m) and x ≡ b (mod n), gcd(m,n) only needs to divide b−a. If it does not, the page identifies the conflicting merge step.'],

    ['精确矩阵实验室 | Dezhonger Tools', 'Exact matrix lab | Dezhonger Tools'],
    ['精确矩阵实验室', 'Exact matrix lab'],
    ['从输入到消元全部使用自动约分的有理数。特征值保留为根式或特征多项式的根，QR 中的平方根也保持符号形式。', 'Every step uses automatically reduced rational numbers. Eigenvalues remain radicals or roots of the characteristic polynomial, and QR square roots stay symbolic.'],
    ['矩阵输入与计算', 'Matrix input and calculation'],
    ['空格或逗号分列，换行或分号分行；最大 10 × 10', 'Separate columns with spaces or commas and rows with line breaks or semicolons; maximum 10 × 10'],
    ['矩阵 A', 'Matrix A'],
    ['完整分析', 'Full analysis'],
    ['对称矩阵示例', 'Symmetric example'],
    ['清空', 'Clear'],
    ['完整分析包含：', 'Full analysis includes:'],
    ['维度、秩、转置、RREF、零空间、列空间；方阵额外计算行列式、迹、逆矩阵、特征多项式、特征值和特征向量/特征空间。', 'Dimensions, rank, transpose, RREF, null space, and column space; square matrices also include determinant, trace, inverse, characteristic polynomial, eigenvalues, and eigenspaces.'],
    ['矩阵分解', 'Matrix decomposition'],
    ['PLU（带行置换）', 'PLU (with row permutation)'],
    ['LDU（单位三角）', 'LDU (unit triangular)'],
    ['LDLᵀ（对称矩阵）', 'LDLᵀ (symmetric matrix)'],
    ['QR（满列秩）', 'QR (full column rank)'],
    ['SVD（精确谱表示）', 'SVD (exact spectral form)'],
    ['执行分解', 'Run decomposition'],

    ['代码注释去除器 | Dezhonger', 'Code comment remover | Dezhonger'],
    ['清理多种语言的代码注释，保留可控的空白处理，并随时对比原文和结果。', 'Remove comments from multiple languages with controlled whitespace handling, then compare the source and result at any time.'],
    ['打开文本对比', 'Open text diff'],
    ['返回首页', 'Back to home'],
    ['多语言注释去除工具', 'Multi-language comment remover'],
    ['复制结果后会自动带着原文和处理结果打开 diff 页。', 'After copying, the Diff page opens with both the source and processed result.'],
    ['语言', 'Language'],
    ['去除行尾空格', 'Trim trailing whitespace'],
    ['去除空行', 'Remove blank lines'],
    ['删除由注释产生的空行', 'Remove lines emptied by comments'],
    ['保留整行注释', 'Keep full-line comments'],
    ['保留块注释', 'Keep block comments'],
    ['替换 long long→LL/ULL', 'Replace long long→LL/ULL'],
    ['去除注释', 'Remove comments'],
    ['复制结果', 'Copy result'],
    ['清除文本', 'Clear text'],
    ['原始代码', 'Source code'],
    ['去掉注释后的代码', 'Code without comments'],
    ['这个页面依赖 JavaScript 运行，当前环境未启用。', 'This page requires JavaScript, which is disabled in the current environment.'],

    ['文本对比工具 | Dezhonger', 'Text diff | Dezhonger'],
    ['使用 Monaco Diff Editor 对比两段文本，也可以从注释去除器直接带入原文和处理结果。', 'Compare two texts with Monaco Diff Editor, or send the source and result directly from the comment remover.'],
    ['返回注释去除器', 'Back to comment remover'],
    ['文本对比工具（Diff）', 'Text diff tool'],
    ['左侧原始文本，右侧修改后文本，左右都可以直接编辑。', 'The original is on the left and the modified text is on the right; both sides are editable.'],
    ['本工具基于', 'This tool uses'],
    ['实现，用于对比两段文本的差异。', 'to compare differences between two texts.'],
    ['左侧为「原始文本」，右侧为「修改后文本」。', 'The original text appears on the left and the modified text on the right.'],
    ['支持自动换行、行号显示，高亮展示增加、删除、修改的部分。', 'Word wrapping, line numbers, and highlights for additions, deletions, and edits are supported.'],
    ['点击「交换左右」可以快速交换原始/修改文本。', 'Use “Swap sides” to exchange the original and modified text.'],
    ['点击「示例」可自动填入一组示例文本，方便体验效果。', 'Use “Load example” to populate a sample comparison.'],
    ['填充示例', 'Load example'],
    ['交换左右', 'Swap sides'],
    ['复制右侧文本', 'Copy right side'],
    ['原始文本 vs 修改后文本', 'Original vs modified'],

    ['服务器测速 | Dezhonger', 'Server speed test | Dezhonger'],
    ['连续请求服务器的轻量 API，测量当前浏览器到服务器的 HTTPS 往返延迟。结果会受到本地网络、Clash 路由和运营商线路影响。', 'Send sequential requests to a lightweight API to measure HTTPS round-trip latency from this browser. Results are affected by your network, Clash routing, and ISP path.'],
    ['平均延迟', 'Average latency'],
    ['点击开始测速', 'Select start to test the connection'],
    ['开始测速', 'Start test'],
    ['最低', 'Minimum'],
    ['中位数', 'Median'],
    ['抖动', 'Jitter'],
    ['成功请求', 'Successful requests'],
    ['服务器区域', 'Server region'],
    ['香港', 'Hong Kong'],
    ['延迟采样趋势图', 'Latency sample trend'],
    ['本次页面连接', 'This page connection'],
    ['总加载', 'Total load'],
    ['测速使用 12 次顺序请求。它反映浏览器到本服务的应用层往返时间，不等同于 ICMP Ping，也不代表服务器访问 Google 的延迟。', 'The test uses 12 sequential requests. It reflects application-layer round trips from this browser—not ICMP Ping or the server\'s latency to Google.'],

    ['数据库浏览器 | Dezhonger', 'Database browser | Dezhonger'],
    ['数据库', 'Database'],
    ['用户管理', 'User management'],
    ['数据库浏览器', 'Database browser'],
    ['查看 public schema 的表、字段定义和数据。页面只提供读取功能，密码哈希、会话令牌和其他敏感列会自动隐藏。', 'Inspect tables, columns, and rows in the public schema. The page is read-only and automatically masks password hashes, session tokens, and other sensitive columns.'],
    ['正在加载表…', 'Loading tables…'],
    ['选择一张表', 'Select a table'],
    ['查看字段结构和当前数据。', 'Inspect its structure and current rows.'],
    ['刷新', 'Refresh'],
    ['数据', 'Data'],
    ['表结构', 'Structure'],
    ['上一页', 'Previous'],
    ['下一页', 'Next'],
    ['安全说明：此页面要求管理员登录；API 不接受 SQL、不提供新增、修改或删除操作，数据库端口仍未对公网开放。', 'Security: this page requires an administrator session. The API accepts no SQL and exposes no create, update, or delete operations; the database port remains private.'],

    ['用户管理 | Dezhonger', 'User management | Dezhonger'],
    ['修改密码', 'Change password'],
    ['创建账号、分配角色、停用账号或重置临时密码。新建和重置后的密码都要求用户首次登录时修改。', 'Create accounts, assign roles, disable access, or reset temporary passwords. New and reset passwords must be changed on first sign-in.'],
    ['创建用户', 'Create user'],
    ['用户名', 'Username'],
    ['临时密码', 'Temporary password'],
    ['角色', 'Role'],
    ['普通用户', 'User'],
    ['管理员', 'Administrator'],
    ['现有用户', 'Existing users'],

    ['修改密码 | Dezhonger', 'Change password | Dezhonger'],
    ['返回备忘录', 'Back to memo'],
    ['首次登录或管理员重置密码后，必须先设置自己的新密码。', 'After your first sign-in or an administrator reset, you must set a new password.'],
    ['当前密码', 'Current password'],
    ['新密码', 'New password'],
    ['再次输入新密码', 'Confirm new password'],
    ['保存新密码', 'Save new password'],

    ['Dezhonger 的私有跨浏览器 Markdown 备忘录。', 'A private, cross-browser Markdown memo by Dezhonger.'],
    ['在线', 'Online'],
    ['退出', 'Sign out'],
    ['正在打开备忘录', 'Opening memo'],
    ['正在恢复登录状态和本地缓存。', 'Restoring your session and local cache.'],
    ['备忘录服务尚未就绪', 'Memo service is not ready'],
    ['页面已经加载，但同源 API 当前不可用。请稍后刷新或联系管理员。', 'The page loaded, but the same-origin API is unavailable. Refresh later or contact the administrator.'],
    ['确认 Docker Compose 中的', 'Confirm that'],
    ['和', 'and'],
    ['健康。', 'are healthy in Docker Compose.'],
    ['确认 Nginx 可以代理', 'Confirm that Nginx proxies'],
    ['使用管理员创建的账号登录。', 'Sign in with an account created by an administrator.'],
    ['登录备忘录', 'Sign in to memo'],
    ['使用管理员创建的用户名和密码登录。本站不开放公开注册。', 'Sign in with credentials created by an administrator. Public registration is disabled.'],
    ['密码', 'Password'],
    ['登录', 'Sign in'],
    ['未启用', 'Disabled'],
    ['密码只通过 HTTPS 提交给同源 API，并以不可逆哈希保存。', 'Passwords are sent only to the same-origin API over HTTPS and stored as irreversible hashes.'],
    ['备忘录列表', 'Memo list'],
    ['新建', 'New'],
    ['搜索备忘录', 'Search memos'],
    ['搜索标题、正文或标签', 'Search title, content, or tags'],
    ['按分类筛选', 'Filter by category'],
    ['全部分类', 'All categories'],
    ['按状态筛选', 'Filter by status'],
    ['全部状态', 'All statuses'],
    ['收件箱', 'Inbox'],
    ['待处理', 'To do'],
    ['进行中', 'In progress'],
    ['已完成', 'Done'],
    ['已归档', 'Archived'],
    ['排序方式', 'Sort order'],
    ['最近更新', 'Recently updated'],
    ['最近创建', 'Recently created'],
    ['标题排序', 'Sort by title'],
    ['状态排序', 'Sort by status'],
    ['按标签筛选', 'Filter by tag'],
    ['导入 JSON', 'Import JSON'],
    ['导出 JSON', 'Export JSON'],
    ['已同步', 'Synced'],
    ['立即同步', 'Sync now'],
    ['记录下一件事', 'Capture what matters next'],
    ['新建一条备忘录，内容会自动保存并同步到其他浏览器。', 'Create a memo that saves automatically and syncs across browsers.'],
    ['新建备忘录', 'New memo'],
    ['标题', 'Title'],
    ['无标题备忘录', 'Untitled memo'],
    ['标签', 'Tags'],
    ['标签，用逗号分隔', 'Tags, separated by commas'],
    ['分类', 'Category'],
    ['例如：工作', 'For example: Work'],
    ['状态', 'Status'],
    ['编辑模式', 'Editing mode'],
    ['编辑', 'Edit'],
    ['预览', 'Preview'],
    ['置顶', 'Pin'],
    ['删除', 'Delete'],
    ['Markdown 正文', 'Markdown content'],
    ['使用 Markdown 记录想法……', 'Write your thoughts in Markdown…'],
    ['Markdown 预览', 'Markdown preview'],
    ['编辑器加载中…', 'Loading editor…'],

    ['标准化方程', 'Normalized equation'],
    ['判别式', 'Discriminant'],
    ['精确因式分解', 'Exact factorization'],
    ['精确解', 'Exact roots'],
    ['Cardano 不变量', 'Cardano invariants'],
    ['Ferrari 降次参数', 'Ferrari reduction parameters'],
    ['辅助方程与中间量', 'Auxiliary equation and intermediate values'],
    ['恒等式，任意复数都是解。', 'Identity: every complex number is a solution.'],
    ['矛盾方程，无解。', 'Contradictory equation: no solution.'],
    ['最高次项为 0，已按一次方程求解。', 'The leading coefficient is 0, so the equation was solved as linear.'],
    ['判别式为 0，存在二重根。', 'The discriminant is 0, so there is a double root.'],
    ['判别式为完全平方数，两个根均为有理数。', 'The discriminant is a perfect square, so both roots are rational.'],
    ['结果保留为精确根式，没有转换为浮点近似。', 'The result remains an exact radical and was not converted to a floating-point approximation.'],
    ['判别式小于 0，得到一对共轭复根。', 'The discriminant is negative, producing a pair of complex conjugate roots.'],
    ['三个根完全重合。', 'All three roots coincide.'],
    ['三次项系数为 0，已降为二次方程。', 'The cubic coefficient is 0, so the equation was reduced to a quadratic.'],
    ['检测到有理根，先精确因式分解，再求剩余二次因子的根。', 'A rational root was found. The equation was factored exactly before solving the remaining quadratic factor.'],
    ['这是纯三次方程，三个复根由一个精确立方根和三次单位根给出。', 'This is a pure cubic; its three complex roots are expressed using one exact cube root and the cube roots of unity.'],
    ['判别式大于 0，有三个互不相同的实根。Cardano 形式中的中间复数会在最终结果中相消。', 'The discriminant is positive, giving three distinct real roots. Intermediate complex values in Cardano form cancel in the final result.'],
    ['判别式小于 0，有一个实根和一对共轭复根。', 'The discriminant is negative, giving one real root and a complex conjugate pair.'],
    ['判别式为 0，方程存在重根。', 'The discriminant is 0, so the equation has repeated roots.'],
    ['结果保留为精确 Cardano 根式；这里已选择不使 C 为 0 的平方根分支。', 'The result remains in exact Cardano radical form, using a square-root branch that keeps C nonzero.'],
    ['四次项系数为 0，已降为三次方程。', 'The quartic coefficient is 0, so the equation was reduced to a cubic.'],
    ['检测到有理根，先做精确综合除法，再求三次因子的全部根。', 'A rational root was found. Exact synthetic division was applied before solving all roots of the cubic factor.'],
    ['平移后奇次项为 0，方程已化为双二次方程并精确求解。', 'After translation, the odd-degree term is 0; the equation was reduced to a biquadratic and solved exactly.'],
    ['一般四次方程使用 Ferrari 公式求解；所有中间量保持为有理数与根式。立方根取与公式一致且使 W 非零的分支。', 'The general quartic is solved with Ferrari\'s formula. All intermediate values remain rational numbers and radicals, using the cube-root branch consistent with the formula and W ≠ 0.'],
    ['计算完成；结果未转换为浮点小数。', 'Calculation complete; results were not converted to floating-point decimals.'],

    ['判定结果', 'Primality result'],
    ['小于 2', 'Less than 2'],
    ['小质数', 'Small prime'],
    ['64 位整数的确定性 Miller–Rabin', 'Deterministic Miller–Rabin for 64-bit integers'],
    ['多基强伪素数检验', 'Multi-base strong probable-prime test'],
    ['素数计数', 'Prime count'],
    ['最大公因数与最小公倍数', 'Greatest common divisor and least common multiple'],
    ['Bézout 等式', 'Bézout identity'],
    ['模逆元', 'Modular inverse'],
    ['标准化同余方程组', 'Normalized congruence system'],
    ['无解', 'No solution'],
    ['通解', 'General solution'],
    ['在 n < 2^64 范围内使用确定性 Miller–Rabin 基底集合，结论为确定性。', 'For n < 2^64, a deterministic Miller–Rabin base set is used, so the conclusion is deterministic.'],
    ['正在进行试除与 Pollard–Brent 分解…', 'Running trial division and Pollard–Brent factorization…'],
    ['质因数分解要求整数 n ≥ 2', 'Prime factorization requires an integer n ≥ 2'],
    ['分解完成；乘积与原数完全一致。', 'Factorization complete; the product exactly matches the original integer.'],
    ['已返回精确的部分分解；困难余因子在本次浏览器时间预算内未拆开。', 'An exact partial factorization was returned; difficult cofactors were not split within this browser time budget.'],
    ['正在初始化筛表并计算 Lehmer π(x)…', 'Initializing the sieve and calculating Lehmer π(x)…'],
    ['计算完成。', 'Calculation complete.'],
    ['模逆元不存在。', 'The modular inverse does not exist.'],
    ['先检查余数之差能否被最大公因数整除，再在约去公因数后的模数上求逆元。', 'First check whether the difference of remainders is divisible by the GCD, then find an inverse after dividing out that GCD.'],
    ['同余方程组无解。', 'The congruence system has no solution.'],
    ['合并完成；不要求输入模数两两互质。', 'Merge complete; input moduli do not need to be pairwise coprime.'],

    ['基本信息', 'Basic information'],
    ['转置', 'Transpose'],
    ['行最简形（RREF）', 'Reduced row echelon form (RREF)'],
    ['零空间', 'Null space'],
    ['列空间', 'Column space'],
    ['行列式与迹', 'Determinant and trace'],
    ['逆矩阵', 'Inverse matrix'],
    ['特征多项式', 'Characteristic polynomial'],
    ['特征值', 'Eigenvalues'],
    ['特征向量', 'Eigenvectors'],
    ['PLU 分解', 'PLU decomposition'],
    ['置换矩阵 P', 'Permutation matrix P'],
    ['下三角矩阵 L', 'Lower-triangular matrix L'],
    ['上三角矩阵 U', 'Upper-triangular matrix U'],
    ['LDU 分解', 'LDU decomposition'],
    ['P 与 L', 'P and L'],
    ['D 与单位上三角矩阵 U', 'D and unit upper-triangular matrix U'],
    ['LDLᵀ 分解', 'LDLᵀ decomposition'],
    ['L 与 D', 'L and D'],
    ['精确 QR 分解', 'Exact QR decomposition'],
    ['正交矩阵 Q', 'Orthogonal matrix Q'],
    ['上三角矩阵 R', 'Upper-triangular matrix R'],
    ['SVD 的精确定义', 'Exact SVD definition'],
    ['Gram 矩阵', 'Gram matrix'],
    ['奇异值', 'Singular values'],
    ['行列式为 0，矩阵奇异。', 'The determinant is 0, so the matrix is singular.'],
    ['对每个特征值代入 v(λ)；若得到零向量，则直接取 ker(A−λI) 的非零向量。', 'Substitute each eigenvalue into v(λ). If it gives the zero vector, choose a nonzero vector directly from ker(A−λI).'],
    ['次数大于 4 的一般多项式不存在统一根式公式；以特征多项式的根定义特征值是精确表示，不是数值近似。', 'General polynomials above degree 4 have no universal radical formula. Defining eigenvalues as roots of the characteristic polynomial is exact, not numerical approximation.'],
    ['分析完成；所有消元与多项式系数均为精确有理数。', 'Analysis complete; all elimination steps and polynomial coefficients are exact rational numbers.'],
    ['平方根保留为符号表达式，因此 QᵀQ=I 与 A=QR 都是精确等式。', 'Square roots remain symbolic, so QᵀQ=I and A=QR are both exact identities.'],
    ['μᵢ 是 AᵀA 的非负特征值；V 的列为对应特征向量，非零奇异值下 Uᵢ=Avᵢ/σᵢ。该表示对任意次数保持精确。', 'μᵢ are the nonnegative eigenvalues of AᵀA; columns of V are their eigenvectors, and Uᵢ=Avᵢ/σᵢ for nonzero singular values. This representation stays exact at any degree.'],
    ['分解完成。', 'Decomposition complete.'],

    ['暂不支持负数的偶次根', 'Even roots of negative numbers are not supported'],
    ['分母不能为 0', 'The denominator cannot be 0'],
    ['存在空的数值', 'A numeric value is empty'],
    ['不能除以 0', 'Division by 0 is not allowed'],
    ['0 没有倒数', '0 has no reciprocal'],
    ['有理数幂只支持整数指数', 'Rational powers require an integer exponent'],
    ['请输入十进制整数', 'Enter a decimal integer'],
    ['请输入非负整数', 'Enter a nonnegative integer'],
    ['请输入矩阵', 'Enter a matrix'],
    ['矩阵每一行的列数必须一致', 'Every matrix row must have the same number of columns'],
    ['当前浏览器版精确支持 0 ≤ x ≤ 10^13', 'This browser version supports exact results for 0 ≤ x ≤ 10^13'],
    ['请至少输入一个同余式', 'Enter at least one congruence'],
    ['一次最多合并 30 个同余式', 'At most 30 congruences can be merged at once'],
    ['请至少提供一个同余式', 'Provide at least one congruence'],
    ['模数不能为 0', 'The modulus cannot be 0'],
    ['模数必须大于 1', 'The modulus must be greater than 1'],

    ['正在本地保存…', 'Saving locally…'],
    ['等待自动保存…', 'Waiting to autosave…'],
    ['已保存在本地，远程同步失败', 'Saved locally; remote sync failed'],
    ['正在刷新…', 'Refreshing…'],
    ['同步失败，联网后将自动重试。', 'Sync failed and will retry automatically when online.'],
    ['正在使用本地缓存', 'Using local cache'],
    ['无法连接远程存储，当前显示本地缓存。', 'Remote storage is unavailable; showing the local cache.'],
    ['备忘录已删除。', 'Memo deleted.'],
    ['已置顶。', 'Pinned.'],
    ['已取消置顶。', 'Unpinned.'],
    ['导入文件不能超过 5 MB。', 'The import file cannot exceed 5 MB.'],
    ['用户名或密码不正确。', 'The username or password is incorrect.'],
    ['登录状态已失效，请重新登录。', 'Your session expired. Sign in again.'],
    ['正在登录…', 'Signing in…'],
    ['登录成功，正在加载备忘录…', 'Signed in. Loading memos…'],
    ['登录失败，请稍后重试。', 'Sign-in failed. Try again later.'],
    ['仍有内容未同步，暂不退出以避免丢失。', 'Some content is still unsynced. Sign-out was paused to avoid data loss.'],
    ['退出失败，请稍后重试。', 'Sign-out failed. Try again later.'],
    ['同步完成。', 'Sync complete.'],
    ['同步失败，已保留本地内容。', 'Sync failed; local content was preserved.'],
    ['同步失败', 'Sync failed'],
    ['导入失败。', 'Import failed.'],
    ['网络已恢复，正在同步。', 'The network is back. Syncing now.'],
    ['已离线，修改会先保存在当前浏览器。', 'You are offline. Changes will be saved in this browser first.'],
    ['已保存。', 'Saved.'],
    ['初始化失败，请刷新页面后重试。', 'Initialization failed. Refresh the page and try again.'],
    ['无法解析 JSON 文件。', 'Unable to parse the JSON file.'],
    ['文件中没有有效的 notes 数组。', 'The file does not contain a valid notes array.'],

    ['用户名或密码不正确', 'invalid username or password'],
    ['需要登录', 'authentication required'],
    ['需要先修改密码', 'password change required'],
    ['需要管理员权限', 'administrator access required'],
    ['密码必须为 12 到 72 个字节', 'password must contain 12 to 72 bytes'],
    ['新密码必须为 12 到 72 个字节', 'new password must contain 12 to 72 bytes'],
    ['当前密码不正确', 'current password is incorrect'],
    ['用户名已经存在', 'username already exists'],
    ['最后一个启用的管理员不能被停用或降级', 'the last active administrator cannot be disabled or demoted'],
    ['不能停用或降级当前管理员账号', 'you cannot disable or demote your own administrator account'],
    ['数据库暂时不可用', 'database unavailable'],
    ['服务器内部错误', 'internal server error'],
  ])

  const reverseTranslations = new Map(
    Array.from(translations, ([chinese, english]) => [english, chinese]),
  )

  const patterns = [
    [/^请求失败（HTTP (\d+)）$/, 'Request failed (HTTP $1)'],
    [/^(\d+) 个字段 · (\d+) 行 · 只读$/, '$1 columns · $2 rows · read-only'],
    [/^(\d+) 张表$/, '$1 tables'],
    [/^页面总耗时 (.+)$/, 'Total page time $1'],
    [/^(\d+) 次请求失败，结果仅基于成功样本。$/, '$1 requests failed; results use successful samples only.'],
    [/^第 (\d+) 次合并$/, 'Merge $1'],
    [/^第 (\d+) 个同余式与当前合并结果冲突，因此整个方程组无解。$/, 'Congruence $1 conflicts with the current merge, so the system has no solution.'],
    [/^主元列（从 1 开始）：(.+)$/, 'Pivot columns (1-based): $1'],
    [/^特征空间 λ = (.+)$/, 'Eigenspace λ = $1'],
    [/^最多支持 (\d+) 位十进制整数$/, 'At most $1 decimal digits are supported'],
    [/^矩阵最多支持 (\d+) 行、(\d+) 列$/, 'Matrices support at most $1 rows and $2 columns'],
    [/^第 (\d+) 行的模数不能为 0$/, 'The modulus on line $1 cannot be 0'],
    [/^第 (\d+) 行格式无法识别，请使用“余数 mod 模数”或“余数 模数”$/, 'Line $1 is not recognized. Use “remainder mod modulus” or “remainder modulus”.'],
    [/^可被 (.+) 整除。$/, 'Divisible by $1.'],
    [/^底数 (.+) 给出了合数见证。$/, 'Base $1 is a compositeness witness.'],
    [/^区间 \[1, (.+)\] 中共有 (.+) 个质数。结果是整数精确值。$/, 'There are $2 primes in [1, $1]. The result is an exact integer.'],
    [/^已通过 (\d+) 个基底的强伪素数检验。由于 n ≥ 2\^64，页面不会把概率性结果伪装成严格素性证明。$/, 'Passed strong probable-prime tests for $1 bases. Because n ≥ 2^64, the page does not present this probabilistic result as a proof.'],
    [/^分解完成；大因子 (.+) 仅通过概率素性检验。$/, 'Factorization complete; large factors $1 passed only probabilistic primality tests.'],
    [/^无法识别有理数“(.+)”，分数请写成 p\/q$/, 'Unrecognized rational number “$1”; write fractions as p/q'],
    [/^无法识别有理数“(.+)”$/, 'Unrecognized rational number “$1”'],
    [/^这是 A−λI 的第 (\d+)、(\d+) 行叉积。对特征值 λ 代入后若恰为零，改取另外两行的叉积或直接求 ker\(A−λI\)。$/, 'This is the cross product of rows $1 and $2 of A−λI. If substituting an eigenvalue yields zero, use another row pair or solve ker(A−λI) directly.'],
    [/^已导出 (\d+) 条备忘录。$/, 'Exported $1 memos.'],
    [/^已导入 (\d+) 条备忘录。$/, 'Imported $1 memos.'],
    [/^单次最多导入 (\d+) 条备忘录。$/, 'At most $1 memos can be imported at once.'],
    [/^第 (\d+) 条备忘录格式不正确。$/, 'Memo $1 has an invalid format.'],
  ]

  function storedLanguage() {
    const requested = new URLSearchParams(root.location.search).get('lang')
    if (requested === 'en' || requested === 'zh') {
      try {
        root.localStorage.setItem(storageKey, requested)
      } catch {
        // The requested language still applies to the current page.
      }
      return requested
    }
    try {
      return root.localStorage.getItem(storageKey) === 'zh' ? 'zh' : 'en'
    } catch {
      return 'en'
    }
  }

  const language = storedLanguage()

  function t(english, chinese) {
    return language === 'zh' ? chinese : english
  }

  function translateString(value) {
    if (typeof value !== 'string') return value
    if (language === 'zh') return reverseTranslations.get(value) || value
    const exact = translations.get(value)
    if (exact) return exact
    let translated = value
    for (const [pattern, replacement] of patterns) {
      if (pattern.test(translated)) {
        translated = translated.replace(pattern, replacement)
        break
      }
    }
    for (const [source, replacement] of translations) {
      if (source.length >= 8 && translated.includes(source)) {
        translated = translated.split(source).join(replacement)
      }
    }
    return translated
  }

  function translateTextNode(node) {
    const original = node.nodeValue
    const trimmed = original.trim()
    if (!trimmed) return
    const translated = translateString(trimmed)
    if (translated !== trimmed) node.nodeValue = original.replace(trimmed, translated)
  }

  function translateElement(element) {
    if (!(element instanceof Element)) return
    for (const attribute of ['aria-label', 'placeholder', 'title', 'data-loading-text']) {
      if (!element.hasAttribute(attribute)) continue
      element.setAttribute(attribute, translateString(element.getAttribute(attribute)))
    }
    if (element.matches('meta[name="description"]')) {
      element.setAttribute('content', translateString(element.getAttribute('content')))
    }
  }

  function translateDocument() {
    document.documentElement.lang = language === 'zh' ? 'zh-CN' : 'en'
    document.title = translateString(document.title)
    document.querySelectorAll('*').forEach(translateElement)
    const walker = document.createTreeWalker(document.body, NodeFilter.SHOW_TEXT)
    let node
    while ((node = walker.nextNode())) {
      if (node.parentElement?.closest('script, style, textarea, code, pre, [data-i18n-ignore]')) continue
      translateTextNode(node)
    }
    if (language === 'zh') {
      document.querySelectorAll('a[href="/knowledge/"]').forEach((link) => {
        link.setAttribute('href', '/knowledge/zh/')
      })
    }
  }

  function addToggle() {
    const target = document.querySelector('.site-nav, .tool-nav, .memo-topbar-actions')
    if (!target || target.querySelector('.language-toggle')) return
    const button = document.createElement('button')
    button.type = 'button'
    button.className = 'language-toggle'
    button.textContent = language === 'en' ? '中文' : 'English'
    button.setAttribute('aria-label', language === 'en' ? 'Switch to Chinese' : '切换到英文')
    button.addEventListener('click', () => {
      try {
        root.localStorage.setItem(storageKey, language === 'en' ? 'zh' : 'en')
      } catch {
        // A reload still applies the default language when storage is unavailable.
      }
      root.location.reload()
    })
    target.append(button)
  }

  root.DezhongerI18n = { language, t, translateString }
  translateDocument()
  addToggle()
})(window)
