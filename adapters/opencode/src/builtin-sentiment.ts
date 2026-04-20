export interface SentimentResult {
    negative_ratio: number; // 0.0~1.0
    attitude_score: number; // 1~10
    approval_ratio: number; // 0.0~1.0
    source: "builtin";
}

const POSITIVE_KEYWORDS = [
    "谢谢", "好的", "不错", "可以", "完美", "很好", "太棒了", "赞", "优秀",
    "thanks", "good", "great", "perfect", "nice", "awesome", "excellent", "well done",
    "看起来不错", "没问题", "正确", "对的", "就是", "对",
    "correct", "right", "looks good", "lgtm",
];

const NEGATIVE_KEYWORDS = [
    "不对", "错误", "不行", "失败", "糟糕", "有问题", "不行", "别这样",
    "wrong", "bad", "fail", "error", "incorrect", "not working", "broken",
    "再试", "重新", "重来", "不要这样", "改回去",
    "redo", "revert", "try again", "don't do that",
];

export function builtinAnalyze(messages: string[]): SentimentResult {
    if (messages.length === 0) {
        return { negative_ratio: -1, attitude_score: -1, approval_ratio: -1, source: "builtin" };
    }

    const combined = messages.join(" ").toLowerCase();

    let positiveCount = 0;
    let negativeCount = 0;

    for (const kw of POSITIVE_KEYWORDS) {
        const regex = new RegExp(escapeRegex(kw), "gi");
        const matches = combined.match(regex);
        if (matches) positiveCount += matches.length;
    }

    for (const kw of NEGATIVE_KEYWORDS) {
        const regex = new RegExp(escapeRegex(kw), "gi");
        const matches = combined.match(regex);
        if (matches) negativeCount += matches.length;
    }

    const total = positiveCount + negativeCount;

    if (total === 0) {
        return {
            negative_ratio: 0,
            attitude_score: 5,
            approval_ratio: 0.5,
            source: "builtin",
        };
    }

    const negativeRatio = negativeCount / total;
    const attitudeScore = Math.max(1, Math.min(10, 5 + ((positiveCount - negativeCount) / total) * 5));
    const approvalRatio = positiveCount / total;

    return {
        negative_ratio: Math.round(negativeRatio * 100) / 100,
        attitude_score: Math.round(attitudeScore * 10) / 10,
        approval_ratio: Math.round(approvalRatio * 100) / 100,
        source: "builtin",
    };
}

function escapeRegex(str: string): string {
    return str.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}
