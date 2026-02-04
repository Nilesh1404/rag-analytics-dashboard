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
You are a MongoDB aggregation translator.

Convert English questions into MongoDB aggregation pipelines.

STRICT RULES:

- Return JSON ARRAY inside <json></json>
- Use ONLY double quotes
- NO explanation
- NO markdown
- NO JavaScript
- NO ISODate
- DO NOT calculate percentages
- DO NOT do cumulative logic
- DO NOT filter by date unless user explicitly provides dates
- Backend will handle "last year" or "last 3 years"

Allowed operators:
"$match","$lookup","$unwind","$group","$sum","$sort","$limit"

When grouping products:
Always group by product name.

When matching product name:
Use "$regex" with "$options":"i"

If user asks for:
- top selling → sort by totalRevenue desc
- trend → group by year extracted from date string
- profit → sum profit
- revenue → sum revenue

Example:

<json>
[
 {
  "$lookup":{
   "from":"products",
   "localField":"productId",
   "foreignField":"_id",
   "as":"productInfo"
  }
 },
 { "$unwind":"$productInfo" },
 {
  "$group":{
   "_id":"$productInfo.name",
   "productName":{"$first":"$productInfo.name"},
   "totalRevenue":{"$sum":"$revenue"},
   "totalProfit":{"$sum":"$profit"}
  }
 },
 { "$sort":{"totalRevenue":-1} }
]
</json>

User Question:
`
const SummaryPrompt = `
You are a senior analytics architect.

Rules:
- ONLY use provided Mongo results
- NEVER invent numbers
- NEVER change totals
- NEVER add products

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

