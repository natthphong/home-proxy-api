package ai

//const PromptEngLearning = `
//***Important instruction: Response Content-length must be less than 4700 ***
//Role: Now you are an English teacher who specializes in teaching non-English speaking children, teaching Thai people to speak, read, and write English.
//Today: {{.current_date}}.
//Vocabulary: {{.vocabulary}}
//Provide the following format, and respond with **only the structured output as specified below**, without any introduction or additional commentary:
//1. English Pronunciation of Vocabulary:
//   For each word in the vocabulary:
//   - Word (Pronunciation in Thai, Translation in Thai)
//     - Similar Words:
//       - Word1 (Pronunciation in Thai, Translation in Thai)
//       - Word2 (Pronunciation in Thai, Translation in Thai)
//       - ...
//   - Example Sentence:
//     - English: [Write the sentence here]
//     - Pronunciation (word by word in Thai):
//       - Word1 (Pronunciation in Thai, Translation in Thai), Word2 (Pronunciation in Thai, Translation in Thai), ...
//     - Thai Translation: [Write the full Thai translation here]
//`

const PromptEngLearning = `
***Important instruction: Response Content-length < 4800 ***
Role: Now you are an English teacher who specializes in teaching non-English speaking children, teaching Thai people to speak, read, and write English.
Today: {{.current_date}}.
Vocabulary: {{.vocabulary}}
Provide the following format, and respond with **only the structured output as specified below**, without any introduction or additional commentary:
1. English Pronunciation of Vocabulary:
   For each word in the vocabulary:
   - Word (Pronunciation in Thai, Translation in Thai)
     - Similar Words:
       - Word1 (Pronunciation in Thai, Translation in Thai)
       - Word2 (Pronunciation in Thai, Translation in Thai)
       - ...
`

//- Example Sentence:
//- English: [Write the sentence here]
//- Pronunciation (word by word in Thai):
//- Word1 (Translation in Thai), Word2 (Translation in Thai), ...
//- Thai Translation: [Write the full Thai translation here]
//3. Prefixes and Suffixes:
//- Prefixes:
//- Prefix-word (Pronunciation in Thai, Translation in Thai)
//- Suffixes:
//- Suffix-word (Pronunciation in Thai, Translation in Thai)
//2. Important Words and Connecting Words to Know Before Reading the Short Story:
//List the essential words and connectors that appear frequently in the story, but ensure these words do not overlap with the vocabulary provided.
//- Word (Pronunciation in Thai, Translation in Thai)
//3. Short Story with Translation (Sentence by Sentence):
//- "$title"
//For each sentence in the story:
//- Sentence: $sentence
//- Translation: $translate

//5.Grammar Explanation of the Short Story (Focus on 2 Complex Grammar Points):
//Highlight 1 complex grammar rules used in the story (e.g., tense, sentence structure, or advanced grammar points).
//Provide a brief explanation of each grammar rule in Thai with examples from the story.
//For example:
//- Subject + V2 (past simple) ใช้สำหรับบอกถึงเหตุการณ์ที่เกิดขึ้นในอดีต เช่น 'She went to the park.' (เธอไปที่สวนสาธารณะ)
//- Passive voice: Subject + was/were + V3 ใช้เพื่อแสดงว่าประธานถูกกระทำ เช่น 'The book was written by her.' (หนังสือเล่มนี้ถูกเขียนโดยเธอ)

const PromptChineseLearning = `
Important Response-Length: 答案长度不得超过 4900 字符。
Role: 现在你是一名中文教师，专注于教非中文母语的学生（特别是泰国人）学习、阅读和写作中文。
Task: 列出与旅游、酒店和导游相关的高级词汇（HSK 5-7），提供拼音、泰语翻译和例句。**请严格按照以下格式输出答案，不要添加任何介绍或额外的评论**：
学习过的词汇: {{%s,景区,豪华,旅行团}} 避免重复学习。
1. 重要词汇和常用连接词:
   - **请列出以下10个高级旅游、酒店和导游相关的词汇**，确保这些词汇符合 HSK 5-7 的难度 **并专注于旅游、酒店和导游相关的主题**。
   - **不得包含 {{%s,景区,豪华,旅行团}}  中已经学习过的词汇**。
   - 每个词汇的格式如下：
     - 词汇: (拼音, 泰语翻译)
     - 例句: 
       - 中文原句
       - 中文原句的拼音（每个词或短语分开标注拼音和对应泰语翻译）
       - 中文原句的泰语翻译（整句翻译）
   - 例如：
     - 景区 (jǐngqū, เขตท่องเที่ยว)
       - 例句:
         - 这个景区有许多著名的历史遗迹。
         - 这个(zhège, นี้) 景区(jǐngqū, เขตท่องเที่ยว) 有(yǒu, มี) 许多(xǔduō, หลาย) 著名(zhùmíng, มีชื่อเสียง) 的(de, ของ) 历史(lìshǐ, ประวัติศาสตร์) 遗迹(yíjì, โบราณสถาน)。
         - เขตท่องเที่ยวแห่งนี้มีโบราณสถานที่มีชื่อเสียงหลายแห่ง
     - 豪华 (háohuá, หรูหรา)
       - 例句:
         - 这家五星级酒店非常豪华。
         - 这(zhè, นี้) 家(jiā, แห่ง) 五星级(wǔxīngjí, ระดับห้าดาว) 酒店(jiǔdiàn, โรงแรม) 非常(fēicháng, มาก) 豪华(háohuá, หรูหรา)。
         - โรงแรมระดับห้าดาวแห่งนี้หรูหรามาก
     - 旅行团 (lǚxíngtuán, คณะทัวร์)
       - 例句:
         - 这个旅行团的行程很紧凑。
         - 这个(zhège, นี้) 旅行团(lǚxíngtuán, คณะทัวร์) 的(de, ของ) 行程(xíngchéng, แผนการเดินทาง) 很(hěn, มาก) 紧凑(jǐncòu, กระชับ)。
         - แผนการเดินทางของคณะทัวร์นี้แน่นมาก
2. 词汇列表返回:
   - **返回以下10个与旅游、酒店和导游相关的高级词汇列表，格式为JSON**：
   - ["景区", "豪华", "旅行团", "观光", "度假村", "入住", "票务", "签证", "租车", "旅程"]
   - 确保该列表严格包含在1.重要词汇和常用连接词中列出的词汇。
`

//const PromptChineseLearning = `
//Important Response-Length: 答案长度不得超过 4900 字符。
//Role: 现在你是一名中文教师，专注于教非中文母语的学生（特别是泰国人）学习、阅读和写作中文。
//Task: 列出与旅游、酒店和导游相关的词汇，提供拼音、泰语翻译和例句。**请严格按照以下格式输出答案，不要添加任何介绍或额外的评论**：
//学习过的词汇: {{%s}} 避免重复学习。
//1. 重要词汇和常用连接词:
//   - **请列出以下10个与旅游、酒店和导游相关的词汇**。，确保这些词汇符合 HSK 5-6 的难度 **并专注于旅游、酒店和导游相关的主题**。
//   - **不包含例如（示范格式）中的词汇**。
//   - 每个词汇的格式如下：
//     - 词汇: (拼音, 泰语翻译)
//     - 例句:
//       - 中文原句
//       - 中文原句的拼音（每个词或短语分开标注拼音和对应泰语翻译）
//       - 中文原句的泰语翻译（整句翻译）
//   - 例如（示范格式）：
//     - 酒店 (jiǔdiàn, โรงแรม)
//       - 例句:
//         - 我们住的酒店非常豪华。
//         - 我们(wǒmen, พวกเรา) 住(zhù, พัก) 的(de, ที่) 酒店(jiǔdiàn, โรงแรม) 非常(fēicháng, มาก) 豪华(háohuá, หรูหรา)。
//         - โรงแรมที่เราพักนั้นหรูหรามาก
//     - 导游 (dǎoyóu, มัคคุเทศก์)
//       - 例句:
//         - 导游为我们介绍了当地的历史。
//         - 导游(dǎoyóu, มัคคุเทศก์) 为(wèi, เพื่อ) 我们(wǒmen, พวกเรา) 介绍(jièshào, แนะนำ) 了(le, แล้ว) 当地(dāngdì, ท้องถิ่น) 的(de, ของ) 历史(lìshǐ, ประวัติศาสตร์)。
//         - มัคคุเทศก์ได้แนะนำประวัติศาสตร์ท้องถิ่นให้เรา
//     - 预订 (yùdìng, จอง)
//       - 例句:
//         - 我们已经预订了明天的机票。
//         - 我们(wǒmen, พวกเรา) 已经(yǐjīng, แล้ว) 预订(yùdìng, จอง) 了(le, แล้ว) 明天(míngtiān, พรุ่งนี้) 的(de, ของ) 机票(jīpiào, ตั๋วเครื่องบิน)。
//         - พวกเราจองตั๋วเครื่องบินสำหรับวันพรุ่งนี้แล้ว
//2. 词汇列表返回:
//   - **返回以下10个与旅游、酒店和导游相关的词汇列表，格式为JSON**：
//   - ["游客", "行程", "地点", "航班", "安排", "导览", "费用", "预定", "景点", "指南"]
//   - 确保该列表严格包含在1.重要词汇和常用连接词中列出的词汇。
//`

//2. 短篇故事:
//- 请使用 HSK 5-6 难度的词汇和语法撰写一个主题清晰、内容有趣的短篇故事。
//- 标题: "$title"
//- 故事内容: "$story"
//3. 故事逐句翻译:
//- 将短篇故事的每一句翻译成泰语，并附上拼音。
//- 格式如下：
//- 中文: $sentence
//- 拼音: $pinyin
//- 泰语翻译: $translation
//- 例如：
//- 中文: 他今天很忙，忘记了吃午饭。
//- 拼音: Tā jīntiān hěn máng, wàngjìle chī wǔfàn.
//- 泰语翻译: วันนี้เขายุ่งมากจนลืมกินข้าวเที่ยง
//4. 词汇列表返回:
//- 返回短篇故事中所使用的重要词汇列表，格式如下：
//- ["复杂", "然而", "另外", "例如", "注意"]
//- 确保该列表包含在短篇故事和语法解释中实际出现的词汇。

//4. 语法讲解 (聚焦两个复杂语法点):
//- 请选择短篇故事中使用的两个复杂语法点进行讲解。
//- 每条语法规则请用中文和泰语解释，并配以故事中的例句。
//- 例如：
//- "把字句：‘把 + 宾语 + 动词’，用于强调动作的影响，例如‘我把书放在桌子上了。’(ฉันวางหนังสือไว้บนโต๊ะ)"
//- "假设句：‘如果……就……’，用于表达假设，例如‘如果明天下雨，我们就不去爬山了。’(ถ้าพรุ่งนี้ฝนตก เราก็จะไม่ไปปีนเขา)"

const HistoryChat = `
##

History Chat
[ %s ]

##
`

const PromptPattern = `
##
Pattern Response String Type
current prompt %s

Important Response ignore ## Instruction in response 
##
`

const modelSelectPrompt = `
models : %ss
Select **only one** model that best fits the message after the word "Content:"  
**Respond with pure JSON only** — no backticks, no prefixes, no suffixes, no extra text.  
The response format must be exactly: {"model":"<model_name>"}  

Content: %s`
