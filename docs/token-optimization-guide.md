# Token Optimization for Dummies
## A Practical Guide to AI Economics — From Basics to Production

**Version 1.0 | June 2026**  
*Bundled with Tokensense — your local AI cost optimizer*

---

> "The difference between a $5/day AI habit and a $500/day AI habit is not intelligence — it's knowing which model to use for which task."

---

## Table of Contents

1. [What Are Tokens? (And Why Should You Care?)](#1-what-are-tokens)
2. [The Model Tier System — Picking the Right Tool](#2-model-tiers)
3. [Prompt Engineering for Token Efficiency](#3-prompt-engineering)
4. [Context Window Management](#4-context-windows)
5. [Multimodal Costs — Images, Audio, and Video](#5-multimodal)
6. [API Key Patterns and Cost Control](#6-api-keys)
7. [Agents and Workflows — Where Costs Explode](#7-agents-workflows)
8. [AI-Powered Tools (Cursor, Claude Desktop, VS Code, etc.)](#8-ai-tools)
9. [Common Mistakes (and How to Avoid Them)](#9-common-mistakes)
10. [Advanced Patterns — Caching, Batching, and Streaming](#10-advanced)

---

## 1. What Are Tokens?

### The One-Paragraph Explanation

A **token** is roughly 3/4 of a word. "Hello, world!" is 4 tokens. "ChatGPT is transforming software engineering" is 7 tokens. AI models price their API by the number of tokens processed — both input (what you send) and output (what the model generates). Understanding tokens is the single most powerful lever for cost control.

### Token Rules of Thumb

| Content Type | Approximate Tokens |
|---|---|
| 1 average English word | ~1.3 tokens |
| 1 line of code | ~5–15 tokens |
| 1 page of text (500 words) | ~650 tokens |
| 1 short function (20 lines) | ~150 tokens |
| 1 small image (low detail) | 85 tokens (flat rate) |
| 1 large image (high detail) | 1,500–3,000 tokens |
| 1 minute of audio | ~1,600 tokens (Whisper) |

### How Pricing Works

```
Cost = (input_tokens / 1,000,000 × input_price) + (output_tokens / 1,000,000 × output_price)
```

**Example — Claude Sonnet 4.5:**
- Input: $3.00 per million tokens
- Output: $15.00 per million tokens
- A 1,000-token prompt generating 500 tokens of code: `(1000/1M × $3) + (500/1M × $15) = $0.003 + $0.0075 = $0.0105`

**That's about 1 cent per interaction.** Multiply by 500 daily AI calls and you're at $5/day — $150/month.

### Key Insight: Output Tokens Cost 5–10× More

Output tokens are almost always priced higher than input tokens. A model that generates verbose responses costs significantly more than one that is concise. **Shorter, more direct outputs save money.**

---

## 2. The Model Tier System — Picking the Right Tool

### The Three Tiers

```
┌─────────────────────────────────────────────────────────────────┐
│  FAST (Haiku, Flash, Mini)                                      │
│  $0.10–$1.00 / million tokens                                   │
│  Best for: autocomplete, quick Q&A, simple extraction           │
│  Speed: ~1,000 tokens/sec                                       │
├─────────────────────────────────────────────────────────────────┤
│  BALANCED (Sonnet, standard GPT-4)                              │
│  $3–$15 / million tokens                                        │
│  Best for: code generation, debugging, documentation            │
│  Speed: ~200 tokens/sec                                         │
├─────────────────────────────────────────────────────────────────┤
│  PREMIUM (Opus, GPT-4 Turbo, Gemini Pro)                       │
│  $15–$75 / million tokens                                       │
│  Best for: complex reasoning, architecture, security analysis   │
│  Speed: ~80 tokens/sec                                          │
└─────────────────────────────────────────────────────────────────┘
```

### Task-to-Model Mapping

| Task | Ideal Tier | Why |
|---|---|---|
| Code autocomplete | Fast | Pattern matching, not reasoning |
| Unit test generation | Fast | Templated, repetitive |
| Docstring writing | Fast | Structured format |
| Bug fixing (simple) | Balanced | Needs code understanding |
| Code review | Balanced | Context + judgment |
| Debugging complex issues | Balanced/Premium | Requires deep analysis |
| Architecture decisions | Premium | High-stakes reasoning |
| Security analysis | Premium | Subtle vulnerability patterns |
| Chat / Q&A | Fast | Conversational |
| Data extraction (structured) | Fast | Rule-following task |
| Long document summarization | Fast/Balanced | Depends on doc complexity |

### Real-World Cost Comparison

Assuming 10,000 tokens in, 2,000 tokens out per session:

| Model | Cost per Session | Daily (20 sessions) | Monthly |
|---|---|---|---|
| Haiku 3.5 | $0.0027 | $0.054 | $1.62 |
| Sonnet 4.5 | $0.046 | $0.92 | $27.60 |
| Opus 4 | $0.18 | $3.60 | $108.00 |

**Using Haiku for simple tasks instead of Opus saves $106/month per developer.** For a team of 10, that's $1,060/month.

### How to Choose

```
Is the task templated/repetitive? → Fast
Does it need code context? → Balanced
Is it a business-critical decision? → Premium
Are you debugging a known bug type? → Balanced
Are you doing security/architecture work? → Premium
Is it an autocomplete/tab suggestion? → Fast
```

---

## 3. Prompt Engineering for Token Efficiency

### The Golden Rule: Be Specific, Not Verbose

**Bad (27 tokens):**
```
Can you please help me write a function that takes in a list of numbers and 
returns only the ones that are even? I'm using Python.
```

**Good (14 tokens):**
```
Python function: filter even numbers from list
```

Both produce the same code. One costs half as much.

### The 5-Point Efficient Prompt Framework

1. **Role** (optional, 5–10 tokens): Only when it genuinely changes output quality
2. **Task** (5–20 tokens): One clear action verb + noun
3. **Context** (10–50 tokens): Only what the model can't infer
4. **Format** (5–10 tokens): Specify if non-default output is needed
5. **Constraints** (5–10 tokens): Language, length, style if critical

**Example — Code generation:**
```
Python. Write a binary search function. Return only code, no explanation.
Input: sorted list + target value. Output: index or -1.
```
*~25 tokens. Clear, complete, efficient.*

### System Prompt Optimization

System prompts run with every request. Even a 200-token system prompt at 1,000 requests/day costs you 200,000 extra tokens/day.

**Audit your system prompts:**
- Remove boilerplate instructions the model already knows
- Compress multi-paragraph instructions to bullet points
- Move static context to user turn (it's often cached differently)
- Use shorter role descriptions: "Expert Python dev" not "You are an expert Python developer with 20 years of experience who..."

**Before (312 tokens):**
```
You are an extremely helpful and knowledgeable AI assistant with deep expertise 
in software engineering, particularly Python. You always write clean, well-documented,
efficient code that follows best practices and PEP 8 style guidelines. You provide
clear explanations and always consider edge cases. You never write buggy code and
you always test your assumptions. When asked to write code, you write production-ready
code that handles errors gracefully...
```

**After (18 tokens):**
```
Expert Python dev. Clean PEP 8 code. Handle errors. No explanations unless asked.
```

### Few-Shot Examples — Use Them Wisely

Few-shot examples dramatically improve output quality — but they cost tokens. Use them when:
- The output format is non-standard
- The task has domain-specific patterns
- You're getting inconsistent results

Avoid them when:
- The task is standard (function writing, Q&A)
- You're using a premium model (already trained well)
- The few-shot examples are long

**Efficient few-shot:**
```
Extract JSON from text.
Text: "John, 25, engineer" → {"name":"John","age":25,"role":"engineer"}
Text: "Alice, 30, designer" → 
```
*Short, pattern-clear, 35 tokens*

### Chain-of-Thought: Only When Needed

Chain-of-thought ("think step by step") significantly increases output tokens. Use it only for:
- Math/logic problems
- Multi-step reasoning
- When you've seen wrong answers without it

For code generation, debugging, and extraction tasks, it usually adds cost without benefit.

---

## 4. Context Window Management

### The Context Window Trap

Every token in your context window costs money. A 100K token context on Claude Sonnet 4.5 costs:
- Input price: `100,000 / 1M × $3 = $0.30` just to process the context
- At 50 requests/day with 100K context: $15/day in context costs alone

### Smart Context Strategies

**1. Sliding window for long conversations**
Don't pass the entire conversation history. Keep only:
- The system prompt
- Last 5–10 turns
- Any pinned critical context

```python
# Instead of:
messages = full_conversation_history  # 200 turns

# Do:
messages = [system_msg] + last_n_turns(full_history, n=8) + [current_msg]
```

**2. Summarize old context**
When a conversation grows long, summarize earlier turns with a cheap model:
```
[Use Haiku to summarize turns 1-20 into 2 paragraphs]
[Use those 2 paragraphs as context for turns 21+]
```

**3. Retrieval Augmented Generation (RAG) over large codebases**
Instead of putting 50,000 lines of code in context:
- Index your codebase with embeddings
- Retrieve only the 10–20 most relevant files
- Pass those files as context

Cost reduction: 90%+ for large codebase Q&A.

**4. Structured context instead of raw text**
Bad: Paste the entire GitHub issue thread (2,000 tokens)
Good: "Issue: auth login fails with expired JWT. Error: 401. Relevant code: [20 lines]" (200 tokens)

### Context Caching — Your Best Friend for Repeated Context

Claude and Gemini support **prompt caching**. When you send the same prefix repeatedly, you pay much less:

- **Anthropic**: Cached tokens cost 10% of normal input price (90% discount)
- **Google**: Cached tokens cost ~25% of normal price

**When to use caching:**
- Long system prompts that repeat every request
- Large documents being analyzed across multiple queries
- Codebase context for an IDE assistant

```python
# Anthropic prompt caching
response = client.messages.create(
    model="claude-sonnet-4-5",
    system=[{
        "type": "text",
        "text": long_codebase_context,  # Will be cached after first call
        "cache_control": {"type": "ephemeral"}
    }],
    messages=[{"role": "user", "content": "Explain the auth module"}]
)
```

**Cache TTL:** 5 minutes (Anthropic). Re-use cached context within this window for maximum savings.

---

## 5. Multimodal Costs — Images, Audio, and Video

### Image Token Pricing

Images are processed differently from text. Most providers use a **tile-based** pricing model.

**Anthropic Claude:**
- Base cost: 85 tokens (flat)
- Each 512×512 tile: +170 tokens
- A 1024×1024 image: 85 + 4×170 = 765 tokens (~$0.002 on Sonnet)
- A 4096×4096 image: 85 + 64×170 = 10,965 tokens (~$0.033)

**OpenAI GPT-4o:**
- Low detail: 85 tokens flat
- High detail: 85 + 170 per tile (same tile formula)

**Key insight: Image detail level dramatically affects cost.**

### Image Optimization Techniques

**1. Resize before sending:**
```python
from PIL import Image

# Shrink to model's effective resolution (no quality gain beyond this)
img = Image.open("screenshot.png")
img.thumbnail((1568, 1568))  # Claude's max tile boundary
img.save("screenshot_opt.png")
```

**2. Use low-detail mode when appropriate:**
- Logo recognition: low detail (85 tokens)
- Reading a table: high detail needed
- UI screenshot analysis: medium (800×600 resize often sufficient)

**3. Convert to JPEG for photographic content:**
PNG → JPEG at 80% quality: same visual information, 60–70% smaller file = faster upload, same token count

**4. Crop before sending:**
If you need to analyze a specific region, crop first. A 200×200 crop costs 85 tokens instead of 2,000+ for the full 4K image.

### Audio (Whisper API)

- Approximately $0.006/minute
- 1 hour of audio: ~$0.36
- Optimization: Trim silence, transcribe at 1x speed, use smaller Whisper model for simple content

### Video

Video is not natively supported by most models. Standard approach:
1. Extract 1 frame per second (or per scene change)
2. Send frames as images
3. Cost = number of frames × image token cost

**Optimization:** Use scene detection to extract keyframes only (often 5–10 frames for a 1-minute video vs 60 frames at 1fps).

---

## 6. API Key Patterns and Cost Control

### The Principle of Least Expensive Model

Structure your application so each task uses the cheapest model that can handle it:

```python
MODEL_ROUTING = {
    "autocomplete":     "claude-haiku-3-5",     # Fast, cheap
    "code_generation":  "claude-sonnet-4-5",     # Balanced
    "code_review":      "claude-sonnet-4-5",     # Balanced  
    "debugging":        "claude-sonnet-4-5",     # Balanced
    "architecture":     "claude-opus-4",         # Premium only when needed
    "chat":             "claude-haiku-3-5",      # Fast for conversational
    "summarization":    "claude-haiku-3-5",      # Fast for structured tasks
}

def get_model(task_type):
    return MODEL_ROUTING.get(task_type, "claude-sonnet-4-5")
```

### Per-User/Per-Team Cost Attribution

Always attach metadata to API calls for cost tracking:

```python
response = client.messages.create(
    model="claude-sonnet-4-5",
    messages=[...],
    metadata={
        "user_id": "user_123",
        "team": "engineering",
        "feature": "code_review",
        "session_id": "sess_abc"
    }
)
# Log: response.usage.input_tokens + output_tokens + metadata
```

### Rate Limiting and Budget Guards

```python
# Simple token budget per user per day
class TokenBudget:
    def __init__(self, user_id, daily_limit=100_000):
        self.user_id = user_id
        self.daily_limit = daily_limit
    
    def check(self, estimated_tokens):
        used = self.get_today_usage()  # from your DB
        if used + estimated_tokens > self.daily_limit:
            raise Exception(f"Daily token budget exceeded ({used}/{self.daily_limit})")
    
    def record(self, tokens_used):
        self.increment_today_usage(tokens_used)
```

### API Key Security Best Practices

1. **Never commit API keys** — use environment variables or secrets managers
2. **Per-environment keys** — different keys for dev/staging/prod
3. **Rotate regularly** — monthly or after any suspected exposure
4. **Monitor usage** — set up billing alerts at 50%, 80%, 100% of budget
5. **Use scoped keys** — if the provider allows read-only or model-specific keys, use them
6. **Vault storage** — HashiCorp Vault, AWS Secrets Manager, or `.env` files (never committed)

```bash
# .env file (add to .gitignore)
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
GEMINI_API_KEY=AIza...

# Load in Python
from dotenv import load_dotenv
load_dotenv()
```

### Billing Alerts

Set alerts at multiple thresholds:

| Provider | Where to set alerts |
|---|---|
| Anthropic | Console → Settings → Usage Limits |
| OpenAI | Platform → Settings → Billing → Usage Limits |
| Google Cloud | Billing → Budgets & Alerts |
| AWS Bedrock | CloudWatch → Billing Alarms |

**Recommended thresholds:** 25%, 50%, 75%, 90% of monthly budget.

---

## 7. Agents and Workflows — Where Costs Explode

### Why Agents Are Expensive

A simple agent loop looks cheap but compounds quickly:

```
Agent iteration 1: 2,000 tokens
→ Tool call result: 500 tokens  
Agent iteration 2: 2,500 tokens (includes previous context)
→ Tool call result: 800 tokens
Agent iteration 3: 3,300 tokens (grows with history)
...
10 iterations: ~30,000+ tokens total
```

A 10-step agent workflow on Sonnet 4.5: **~$0.45**. Run 100 agents per day: **$45/day**.

### Agent Cost Reduction Strategies

**1. Limit iterations hard:**
```python
agent_config = {
    "max_iterations": 10,   # Never let agents run forever
    "max_tokens_per_step": 2000,
    "total_budget_tokens": 20_000
}
```

**2. Use fast models for tool-calling steps:**
```python
# Planning step: premium model (complex reasoning)
plan = call_model("claude-opus-4", task_description)

# Execution steps: balanced model (following the plan)
for step in plan.steps:
    result = call_model("claude-haiku-3-5", step)
```

**3. Compress tool results before feeding back:**
```python
# Raw API response: 5,000 tokens
raw_result = search_api(query)

# Compress before feeding to agent: 200 tokens
compressed = extract_key_fields(raw_result)
# OR use a cheap model to summarize: ~100 tokens
compressed = haiku.summarize(raw_result, max_tokens=200)
```

**4. Clear history between task phases:**
```python
# Phase 1: Research (2,000 tokens)
research_result = agent.run("research the topic", history=[])

# Phase 2: Write (start fresh with only the summary)
summary = compress(research_result)
final = agent.run("write the report", history=[
    {"role": "user", "content": f"Research summary: {summary}"}
])
# Cost: 2× single-phase cost instead of 4× cumulative cost
```

### Multi-Agent Systems

For multi-agent workflows, the costs multiply:

```
Orchestrator: 5,000 tokens
  → Agent A: 3,000 tokens
  → Agent B: 4,000 tokens  
  → Agent C: 2,500 tokens
Synthesis: 8,000 tokens (includes all sub-results)
Total: ~22,500 tokens per run
```

**Cost controls:**
- Use fast/balanced models for leaf agents
- Use premium model only for orchestration
- Set total budget limits, not just per-agent limits
- Cache static context shared across all agents

### Tool Use (Function Calling) Costs

Every tool definition you include in your API call costs tokens:

```json
// Each tool definition costs ~50-200 tokens in system tokens
{
    "name": "search_web",
    "description": "Search the internet for recent information",
    "input_schema": {
        "type": "object",
        "properties": {
            "query": {"type": "string"}
        }
    }
}
```

**Best practice:** Only include tools that might be needed for the specific task. Don't dump your full tool library into every call.

### Streaming vs Non-Streaming

For cost, streaming and non-streaming produce identical token counts. Streaming benefits:
- Better UX (user sees output sooner)
- Ability to cancel mid-response (stop paying for tokens you don't need)
- Better detection of errors before the full response arrives

**Cost tip:** With streaming, you can implement client-side stopping:
```python
with client.messages.stream(model=..., messages=...) as stream:
    for text in stream.text_stream:
        yield text
        if stop_condition_met():
            break  # Saves remaining output tokens
```

---

## 8. AI-Powered Tools (Cursor, Claude Desktop, VS Code, Windsurf)

### Understanding Tool Usage Patterns

Different AI coding tools have vastly different token consumption patterns:

| Tool | Average tokens/session | Primary cost driver |
|---|---|---|
| Cursor | 2,000–15,000 | Codebase indexing + chat |
| GitHub Copilot | 200–500 | Inline suggestions |
| Claude Desktop | 1,000–50,000 | Depends heavily on usage |
| VS Code + Copilot | 500–3,000 | Completions + chat |
| Windsurf | 2,000–20,000 | Agent mode usage |

### Cursor Cost Optimization

**What drives cost in Cursor:**
1. `@codebase` — indexes and retrieves from your entire codebase (expensive)
2. Long chat histories — context accumulates across turns
3. Model selection — Claude Opus in Cursor costs 10× Haiku

**Tips:**
- Use `@file` instead of `@codebase` when you know which file
- Use `@symbol` for specific functions
- Reset conversation when changing tasks (Cmd+N)
- Set Cursor to use Haiku/Sonnet for completions, Opus only for complex chat

### Claude Desktop Cost Optimization

- Claude Desktop uses cert pinning — Tokensense sees it in metadata-only mode
- Monitor usage via Claude.ai usage dashboard
- Use Projects for persistent system prompts (set once, amortized)
- Avoid uploading large files repeatedly — attach once per project

### VS Code Copilot / Codeium / Continue

- These typically handle context on their side — you see per-completion costs
- For API-based tools (Continue, etc.), configure model routing
- Use local models (Ollama) for basic completions — $0 cost

### When to Use Local Models

Local models (Ollama, LM Studio) have zero per-token cost after setup:
- CPU/GPU power replaces API costs
- Good for: simple completions, privacy-sensitive code, offline use
- Not good for: complex reasoning, long-context tasks, multimodal

**Cost comparison (per developer/month):**
- Ollama (Llama 3 7B) for completions + GPT-4o for complex: ~$30/month
- Pure cloud premium models: ~$150-400/month
- Hybrid approach: **70-80% savings**

---

## 9. Common Mistakes (and How to Avoid Them)

### Mistake #1: Using Premium Models for Everything

**The problem:** Using Claude Opus or GPT-4 for autocomplete, chat, and simple Q&A.

**The cost:** 10–50× more expensive than necessary.

**The fix:** Route by task type. Use Tokensense to see which tasks are consuming premium models unnecessarily.

```
tokensense ask "write a unit test for my login function"
# → Recommends Haiku (not Opus) — 15× cheaper, same quality
```

---

### Mistake #2: Repeating Context in Every Request

**The problem:** Sending the same 5,000-token codebase context with every single request.

**The cost:** 5,000 tokens × $3/M = $0.015 per request. At 200 requests/day: $3/day = $90/month just in repeated context.

**The fix:** Use prompt caching (Anthropic, Gemini) or structure your calls to reuse the same cache prefix.

---

### Mistake #3: Not Setting Max Tokens on Output

**The problem:** Allowing unbounded output generation.

**The cost:** A model asked to "explain this codebase" might generate 10,000 tokens when 500 would suffice.

**The fix:**
```python
response = client.messages.create(
    model="claude-sonnet-4-5",
    max_tokens=1000,  # Always set this!
    messages=[...]
)
```

---

### Mistake #4: Verbose System Prompts

Already covered in Section 3, but it bears repeating: every 100 tokens in your system prompt costs you 100 tokens per request. At 1,000 requests/day, a 500-token system prompt costs 500,000 extra tokens/day.

---

### Mistake #5: Not Monitoring Agent Costs

**The problem:** Setting up an agent and letting it run for hours without budget guards.

**The cost:** An agent that makes 100 iterations at 3,000 tokens each = 300,000 tokens = $0.90 on Sonnet. But if it gets stuck in a loop...

**The fix:** Always implement iteration limits, token budgets, and logging.

---

### Mistake #6: Over-Engineering Prompts with Chain-of-Thought

**The problem:** Adding "think step by step" to every prompt.

**The cost:** CoT can triple output length.

**The fix:** Only use CoT for tasks that genuinely require multi-step reasoning (math, logic puzzles, complex tradeoffs). For code generation and data extraction, it just adds cost.

---

### Mistake #7: Sending Uncompressed Images

**The problem:** Sending raw screenshots or unresized images.

**The cost:** A 4K monitor screenshot (3840×2160) sent as PNG costs ~16,000 tokens. The same screenshot resized to 1280×720 costs ~2,000 tokens.

**The fix:** Always resize images to the minimum resolution needed for the task before sending.

---

### Mistake #8: Ignoring Model Context Window Waste

**The problem:** Sending 50,000 tokens of context but only 500 tokens are relevant to the question.

**The fix:** Use RAG, summarization, or selective file inclusion rather than dumping everything in context.

---

### Mistake #9: Not Using the Right Pricing Tier for Batch Jobs

**The problem:** Running large batch processing jobs using the synchronous API (full price).

**The fix:** Most providers offer batch/async APIs at 50% discount:
- **Anthropic Message Batches API**: 50% off, 24-hour window
- **OpenAI Batch API**: 50% off, async processing

For non-real-time tasks (nightly reports, bulk analysis, dataset processing), always use the batch API.

---

### Mistake #10: No Cost Attribution = No Visibility

**The problem:** Running AI calls without tracking which feature, user, or team is spending what.

**The consequence:** You get a $3,000 API bill with no idea where it came from.

**The fix:** Instrument every AI call with metadata from day one. Tokensense does this automatically for intercepted calls.

---

## 10. Advanced Patterns — Caching, Batching, and Streaming

### Prompt Caching Deep Dive

**Anthropic's caching rules:**
- Cache prefix must be 1,024+ tokens (Sonnet/Haiku) or 2,048+ (Opus)
- Cache TTL: 5 minutes (base), extendable
- Cost: 25% to write the cache, 10% to read
- Savings kick in after the 1st request

**Optimal caching architecture:**
```
[Cacheable prefix: system prompt + static context]
         ↓ (cached — pay 10% on subsequent requests)
[Dynamic suffix: user's question]
         ↓ (not cached — pay 100%)
[Model response]
```

**Break-even point:** If your static prefix is 4,000 tokens, caching saves money after just 2 requests.

### Batch API Usage

```python
# Anthropic Message Batches — 50% cost reduction
import anthropic

client = anthropic.Anthropic()

# Create a batch of requests
batch = client.messages.batches.create(
    requests=[
        {
            "custom_id": f"task_{i}",
            "params": {
                "model": "claude-haiku-3-5",
                "max_tokens": 500,
                "messages": [{"role": "user", "content": tasks[i]}]
            }
        }
        for i in range(len(tasks))
    ]
)

# Poll for completion (up to 24 hours)
while batch.processing_status == "in_progress":
    time.sleep(60)
    batch = client.messages.batches.retrieve(batch.id)
```

### Speculative Decoding (Provider-Side Optimization)

Some providers use speculative decoding internally — you don't control it, but you benefit from faster output at the same price. Shorter output lengths benefit more.

### Cost Monitoring Setup

**Minimum viable monitoring:**
```python
class AICallLogger:
    def log(self, model, task_type, input_tokens, output_tokens, cost_usd, latency_ms):
        # Write to your database or analytics
        event = {
            "timestamp": datetime.utcnow().isoformat(),
            "model": model,
            "task_type": task_type,
            "tokens_in": input_tokens,
            "tokens_out": output_tokens,
            "cost_usd": cost_usd,
            "latency_ms": latency_ms
        }
        self.db.insert(event)
```

Or just use **Tokensense** — it does this automatically for all AI tools.

### The Optimization Flywheel

```
1. Measure      → Tokensense proxy intercepts all AI calls
2. Classify     → Each call categorized by task type
3. Analyze      → Daily report shows cost by task and tool
4. Identify     → "We're using Opus for test generation — Haiku is fine"
5. Route        → Update model selection for that task type
6. Measure      → See the savings in next day's report
7. Repeat       → Each iteration finds the next optimization
```

---

## Quick Reference Cheatsheet

### Token Estimation
- 1 word ≈ 1.3 tokens
- 1 code line ≈ 8 tokens  
- 1 page ≈ 650 tokens
- 1 image (1024px) ≈ 765 tokens (Anthropic)

### Model Selection
- Simple tasks → Fast (Haiku, Flash, Mini)
- Code/debug → Balanced (Sonnet, GPT-4o)
- Architecture/security → Premium (Opus, Gemini Pro)

### Top 5 Money-Savers
1. Route by task type (biggest impact, up to 95% savings per task)
2. Enable prompt caching for repeated system prompts (90% discount on cached tokens)
3. Use batch API for non-real-time work (50% discount)
4. Resize images before sending (up to 90% reduction)
5. Set max_tokens on all calls (prevent runaway output)

### Emergency Cost Cuts
1. Switch all non-critical tasks to fast tier immediately
2. Enable metadata-only mode (no classification needed = faster + free)
3. Add max_tokens=500 to all calls
4. Pause agent workflows and audit iteration counts
5. Check for any runaway loops in your monitoring

---

*This guide is bundled with [Tokensense](https://github.com/dibakshya/tokensense) — the open-source AI token usage optimizer.*  
*Run `tokensense ask "..."` to get model recommendations for any task.*  
*Run `tokensense report` to see your daily cost breakdown.*

---

**Created by Dibakshya Chakraborty — Product Builder**

**License:** MIT | **Issues:** github.com/dibakshya/tokensense/issues | **Contributing:** See CONTRIBUTING.md
