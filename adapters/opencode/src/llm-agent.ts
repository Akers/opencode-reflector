// eslint-disable-next-line @typescript-eslint/no-explicit-any
type OpencodeClient = any;

export interface ModelInfo {
    providerID: string;
    modelID: string;
}

export interface LLMResult {
    text: string;
    tokens: {
        prompt: number;
        completion: number;
    };
}

export class OpencodeLLMAgent {
    private client: OpencodeClient;
    private cachedModel: ModelInfo | null = null;

    constructor(client: OpencodeClient) {
        this.client = client;
    }

    async resolveModel(override?: string): Promise<ModelInfo> {
        if (override && override.includes("/")) {
            const [providerID, modelID] = override.split("/", 2);
            return { providerID, modelID };
        }

        if (this.cachedModel) {
            return this.cachedModel;
        }

        try {
            const configResp = await this.client.config.get();
            const modelStr = configResp.data?.model;
            if (modelStr && modelStr.includes("/")) {
                const [providerID, modelID] = modelStr.split("/", 2);
                this.cachedModel = { providerID, modelID };
                return this.cachedModel;
            }
        } catch (err) {
            console.warn("[reflector] Failed to get model from SDK:", err);
        }

        return { providerID: "minimax", modelID: "MiniMax-M2.7-highspeed" };
    }

    async callLLM(
        systemPrompt: string,
        userContent: string,
        modelOverride?: string,
    ): Promise<LLMResult> {
        const model = await this.resolveModel(modelOverride);

        const createResult = await this.client.session.create({
            body: { title: `__reflector_llm_${Date.now()}` },
        });
        const sessionId = createResult.data!.id;

        try {
            const promptResult = await this.client.session.prompt({
                path: { id: sessionId },
                body: {
                    system: systemPrompt,
                    model,
                    tools: {},
                    parts: [{ type: "text" as const, text: userContent }],
                },
            });

            const parts = promptResult.data?.parts ?? [];
            const textPart = parts.find((p: any) => p.type === "text");
            const text = textPart && "text" in textPart ? (textPart as any).text : "";

            const info = promptResult.data?.info;
            const tokens = {
                prompt: info?.tokens?.input ?? 0,
                completion: info?.tokens?.output ?? 0,
            };

            return { text, tokens };
        } finally {
            try {
                await this.client.session.delete({ path: { id: sessionId } });
            } catch {
                // best effort
            }
        }
    }
}
