package prompts

const SchemaPrompt = `
You know this MongoDB schema.

Collection: sales
Fields:
- productId (string)
- date (date)
- quantity (number)
- revenue (number)
- profit (number)

Collection: products
Fields:
- _id (string)
- name (string)
- category (string)
- price (number)
- profitMargin (number)

Join rule:
sales.productId == products._id

Each sales document represents one transaction on one date.

Revenue and profit analytics are computed by aggregating sales.

Cumulative percentages (70%, 60%) are NOT done in MongoDB.
They are handled by backend code.
Dates are real MongoDB date objects.
Year-based trends must use $year operator.
`
const IntentPrompt = `
You are a classifier.

If the user asks about:
- revenue
- profit
- sales
- products
- quantities
- trends
- top selling

Then return analytics.

Otherwise return chat.

Return STRICT JSON inside <json></json> only.

<json>
{ "type":"analytics|chat" }
</json>

User Question:
`
const PipelinePrompt = `
You are a MongoDB aggregation expert.

Convert English analytics questions into MongoDB aggregation pipelines.

STRICT RULES:

- Return ONLY JSON ARRAY inside <json></json>
- Use double quotes only
- No explanations
- No markdown
- No JavaScript
- No ISODate
- Do NOT calculate percentages or cumulative logic
- Backend handles last N years filtering

Allowed operators:
"$match","$lookup","$unwind","$group","$sum","$sort","$limit","$addToSet","$project"

Rules:

- Always $lookup products
- Always $unwind productInfo
- When grouping products, group by product name
- For revenue use "$sum":"$revenue"
- For profit use "$sum":"$profit"
- For trends use:

"_id":{"year":{"$year":"$date"}}

- For counting products:

"productCount":{"$addToSet":"$productInfo.name"}

Examples:

Revenue trend:

<json>
[
 {"$lookup":{"from":"products","localField":"productId","foreignField":"_id","as":"productInfo"}},
 {"$unwind":"$productInfo"},
 {"$group":{"_id":"$productInfo.name","totalRevenue":{"$sum":"$revenue"}}},
 {"$project":{"_id":0,"productName":"$_id","revenue":"$totalRevenue"}},
 {"$sort":{"revenue":-1}}
]
</json>

User Question:
`

const SummaryPrompt = `
You are a senior analytics architect.

Rules:
- ONLY use provided Mongo results
- NEVER invent numbers
- If question requires cumulative logic, say backend will handle it.
- Explain results clearly.

You MUST embed ALL product names and numeric values directly into imagePrompt.

Return STRICT JSON inside <json></json>.

Format:

<json>
{
 "summary":"short business explanation using Mongo values",
 "chart":"pie|bar|line",
 "imagePrompt":"Real BI dashboard showing EXACT MongoDB data: list each product with revenue and profit, labeled pie chart with percentages, bar chart with product vs revenue, line chart trend, KPI cards showing totals. MUST include these values explicitly:",
 "videoPrompt":"optional"
}
</json>

Mongo Results:
`

