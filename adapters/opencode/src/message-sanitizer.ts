const API_KEY_PATTERN = /sk-[a-zA-Z0-9]{20,}/g;
const BEARER_TOKEN_PATTERN = /Bearer [a-zA-Z0-9\-._~+/]+=*/g;
const PASSWORD_URL_PATTERN = /:\/\/[^:]+:[^@]+@/g;

function makeGenericPasswordPattern(): RegExp {
    return new RegExp("(password|passwd|pwd)\\s*[=:]\\s*\\S+", "gi");
}

export function sanitizeMessages(messages: string[]): string[] {
    return messages.map((msg) => {
        let sanitized = msg;
        sanitized = sanitized.replace(API_KEY_PATTERN, "[REDACTED_API_KEY]");
        sanitized = sanitized.replace(BEARER_TOKEN_PATTERN, "Bearer [REDACTED]");
        sanitized = sanitized.replace(PASSWORD_URL_PATTERN, "://[REDACTED]:[REDACTED]@");
        const genericPasswordPattern = makeGenericPasswordPattern();
        sanitized = sanitized.replace(genericPasswordPattern, "$1=[REDACTED]");
        return sanitized;
    });
}
