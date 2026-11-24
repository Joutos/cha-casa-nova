import { nextStep } from "@/contexts/RsvpFlowContext";

describe("nextStep", () => {
  it("segue a ordem presença → contribuição → mensagem, pulando o que não foi escolhido", () => {
    const state = { choices: ["attend", "message"] as const, completed: [] };
    expect(nextStep({ ...state, choices: [...state.choices] })).toBe("attend");
    expect(nextStep({ choices: ["attend", "message"], completed: ["attend"] }, "attend")).toBe("message");
  });

  it("retorna null quando tudo foi concluído", () => {
    expect(nextStep({ choices: ["contribute"], completed: ["contribute"] }, "contribute")).toBeNull();
    expect(nextStep({ choices: [], completed: [] })).toBeNull();
  });
});
