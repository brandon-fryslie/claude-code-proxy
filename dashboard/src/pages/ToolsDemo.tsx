import { type FC } from 'react';
import {
  ToolUseContainer,
  ToolResultContent,
} from '@/components/ui';

/**
 * Demo page showcasing all tool components
 * This demonstrates Phase 3 implementation
 */
export const ToolsDemo: FC = () => {
  return (
    <div className="p-8 space-y-8 max-w-6xl mx-auto">
      <div>
        <h1 className="text-3xl font-bold text-gray-900 mb-2">Phase 3: Tool Support Demo</h1>
        <p className="text-gray-600">
          Demonstrating specialized tool renderers and enhanced content detection
        </p>
      </div>

      {/* Bash Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Bash Tool</h2>
        <ToolUseContainer
          id="toolu_01AbCdEfGhIjKlMnOpQr"
          name="Bash"
          defaultExpanded={true}
          input={{
            command: 'npm run build && npm test',
            description: 'Build and test the application',
            timeout: 30000,
          }}
        />
      </section>

      {/* Read Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Read Tool</h2>
        <ToolUseContainer
          id="toolu_02AbCdEfGhIjKlMnOpQr"
          name="Read"
          defaultExpanded={true}
          input={{
            file_path: '/Users/bmf/code/project/src/components/Button.tsx',
            offset: 10,
            limit: 50,
          }}
        />
      </section>

      {/* Write Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Write Tool</h2>
        <ToolUseContainer
          id="toolu_03AbCdEfGhIjKlMnOpQr"
          name="Write"
          defaultExpanded={true}
          input={{
            file_path: '/Users/bmf/code/project/src/utils/helpers.ts',
            content: `export function formatDate(date: Date): string {
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  });
}

export function capitalize(str: string): string {
  return str.charAt(0).toUpperCase() + str.slice(1);
}`,
          }}
        />
      </section>

      {/* Edit Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Edit Tool (Side-by-Side Diff)</h2>
        <ToolUseContainer
          id="toolu_04AbCdEfGhIjKlMnOpQr"
          name="Edit"
          defaultExpanded={true}
          input={{
            file_path: '/Users/bmf/code/project/src/App.tsx',
            old_string: `function App() {
  const [count, setCount] = useState(0);

  return (
    <div>
      <h1>Counter: {count}</h1>
      <button onClick={() => setCount(count + 1)}>
        Increment
      </button>
    </div>
  );
}`,
            new_string: `function App() {
  const [count, setCount] = useState(0);
  const [disabled, setDisabled] = useState(false);

  return (
    <div>
      <h1>Counter: {count}</h1>
      <button
        onClick={() => setCount(count + 1)}
        disabled={disabled}
      >
        Increment
      </button>
      <button onClick={() => setDisabled(!disabled)}>
        Toggle Disabled
      </button>
    </div>
  );
}`,
          }}
        />
      </section>

      {/* Glob Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Glob Tool</h2>
        <ToolUseContainer
          id="toolu_05AbCdEfGhIjKlMnOpQr"
          name="Glob"
          defaultExpanded={true}
          input={{
            pattern: '**/*.tsx',
            path: '/Users/bmf/code/project/src',
          }}
        />
      </section>

      {/* Grep Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Grep Tool</h2>
        <ToolUseContainer
          id="toolu_06AbCdEfGhIjKlMnOpQr"
          name="Grep"
          defaultExpanded={true}
          input={{
            pattern: 'useState|useEffect',
            path: '/Users/bmf/code/project/src',
            type: 'tsx',
            output_mode: 'content',
          }}
        />
      </section>

      {/* Task Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Task Tool (Sub-agent)</h2>
        <ToolUseContainer
          id="toolu_07AbCdEfGhIjKlMnOpQr"
          name="Task"
          defaultExpanded={true}
          input={{
            subagent_type: 'code-reviewer',
            description: 'Review the authentication implementation',
            prompt: 'Please review the authentication code for security vulnerabilities...',
            model: 'gpt-4o',
          }}
        />
      </section>

      {/* TodoWrite Tool */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">TodoWrite Tool</h2>
        <ToolUseContainer
          id="toolu_08AbCdEfGhIjKlMnOpQr"
          name="TodoWrite"
          defaultExpanded={true}
          input={{
            todos: [
              {
                content: 'Implement user authentication',
                status: 'completed',
                priority: 'high',
              },
              {
                content: 'Add password reset functionality',
                status: 'in_progress',
                activeForm: 'Writing reset email template',
                priority: 'high',
              },
              {
                content: 'Set up OAuth providers',
                status: 'pending',
                priority: 'medium',
              },
              {
                content: 'Add session management',
                status: 'pending',
                priority: 'medium',
              },
              {
                content: 'Write authentication tests',
                status: 'pending',
                priority: 'low',
              },
            ],
          }}
        />
      </section>

      {/* Tool Results */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">
          Tool Results (Content Detection)
        </h2>

        <div className="space-y-4">
          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">Success Result</h3>
            <ToolResultContent content="Success" />
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">Error Result</h3>
            <ToolResultContent
              content="Error: ENOENT: no such file or directory, open '/path/to/missing/file.txt'"
              isError={true}
            />
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">JSON Result</h3>
            <ToolResultContent
              content={JSON.stringify({
                status: 'ok',
                data: { users: 42, active: 38 },
                timestamp: '2024-12-25T00:00:00Z',
              })}
            />
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">File List Result</h3>
            <ToolResultContent
              content={`src/components/Button.tsx
src/components/Input.tsx
src/pages/Dashboard.tsx
src/pages/Settings.tsx
src/utils/helpers.ts
src/lib/api.ts`}
            />
          </div>

          <div>
            <h3 className="text-sm font-medium text-gray-700 mb-2">Code Result</h3>
            <ToolResultContent
              content={`function greet(name: string): string {
  return \`Hello, \${name}!\`;
}

export default greet;`}
            />
          </div>
        </div>
      </section>

      {/* Image Content */}
      <section>
        <h2 className="text-xl font-semibold text-gray-800 mb-4">Image Content (with Lightbox)</h2>
        <p className="text-sm text-gray-600 mb-4">
          Click image to open lightbox. (Demo with placeholder - real usage would have base64 data)
        </p>
        <div className="bg-gray-100 p-4 rounded-lg border border-gray-300">
          <p className="text-xs text-gray-500 italic">
            ImageContent component ready - requires base64 image data from actual tool results
          </p>
        </div>
      </section>
    </div>
  );
};

export default ToolsDemo;
