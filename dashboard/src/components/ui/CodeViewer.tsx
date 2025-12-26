import { type FC, useState } from 'react';
import { FileCode, Download, Maximize2, X } from 'lucide-react';
import { CopyButton } from './CopyButton';
import { cn } from '@/lib/utils';

interface CodeViewerProps {
  code: string;
  filename?: string;
  language?: string;
  maxHeight?: number;
  showControls?: boolean;
}

export const CodeViewer: FC<CodeViewerProps> = ({
  code,
  filename,
  language,
  maxHeight = 500,
  showControls = true,
}) => {
  const [isFullscreen, setIsFullscreen] = useState(false);

  // Determine language from file extension
  const getLanguageFromFileName = (fname?: string): string => {
    if (!fname) return 'text';

    const extension = fname.split('.').pop()?.toLowerCase();
    const languageMap: Record<string, string> = {
      js: 'javascript',
      jsx: 'javascript',
      ts: 'typescript',
      tsx: 'typescript',
      py: 'python',
      go: 'go',
      rs: 'rust',
      rb: 'ruby',
      java: 'java',
      c: 'c',
      cpp: 'cpp',
      cs: 'csharp',
      php: 'php',
      swift: 'swift',
      kt: 'kotlin',
      sh: 'bash',
      bash: 'bash',
      sql: 'sql',
      html: 'html',
      css: 'css',
      scss: 'scss',
      json: 'json',
      yaml: 'yaml',
      yml: 'yaml',
      toml: 'toml',
      md: 'markdown',
      xml: 'xml',
    };

    return languageMap[extension || ''] || 'text';
  };

  const detectedLanguage = language || getLanguageFromFileName(filename);

  // Simple syntax highlighting
  const highlightCode = (codeText: string): string => {
    const escapeHtml = (str: string) =>
      str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');

    const tokenPatterns = [
      { regex: /"(?:[^"\\]|\\.)*"/, className: 'text-green-400' },
      { regex: /'(?:[^'\\]|\\.)*'/, className: 'text-green-400' },
      { regex: /`(?:[^`\\]|\\.)*`/, className: 'text-green-400' },
      { regex: /\/\/.*$/, className: 'text-gray-500 italic' },
      { regex: /\/\*[\s\S]*?\*\//, className: 'text-gray-500 italic' },
      { regex: /#.*$/, className: 'text-gray-500 italic' },
      {
        regex:
          /\b(function|const|let|var|if|else|for|while|return|class|import|export|from|async|await|def|elif|except|finally|lambda|with|as|raise|del|global|nonlocal|assert|break|continue|try|catch|throw|new|this|super|extends|implements|interface|abstract|static|public|private|protected|void|int|string|boolean|float|double|char|long|short|byte|enum|struct|typedef|union|namespace|using|package|goto|switch|case|default|fn|pub|mod|use|mut|match|loop|impl|trait|where|type|readonly|override)\b/,
        className: 'text-blue-400',
      },
      {
        regex: /\b(true|false|null|undefined|nil|None|True|False|NULL)\b/,
        className: 'text-orange-400',
      },
      { regex: /\b\d+\.?\d*\b/, className: 'text-purple-400' },
      { regex: /\b[A-Z][a-zA-Z0-9]*\b/, className: 'text-cyan-400' },
    ];

    const combinedPattern = new RegExp(
      tokenPatterns.map((p) => `(${p.regex.source})`).join('|'),
      'gm'
    );

    let result = '';
    let lastIndex = 0;

    for (const match of codeText.matchAll(combinedPattern)) {
      if (match.index! > lastIndex) {
        result += escapeHtml(codeText.slice(lastIndex, match.index));
      }

      const matchedText = match[0];
      let className = '';
      for (let i = 0; i < tokenPatterns.length; i++) {
        if (match[i + 1] !== undefined) {
          className = tokenPatterns[i].className;
          break;
        }
      }

      result += `<span class="${className}">${escapeHtml(matchedText)}</span>`;
      lastIndex = match.index! + matchedText.length;
    }

    if (lastIndex < codeText.length) {
      result += escapeHtml(codeText.slice(lastIndex));
    }

    return result;
  };

  const handleDownload = () => {
    const blob = new Blob([code], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename || 'code.txt';
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  const lines = code.split('\n');
  const lineCount = lines.length;

  const CodeDisplay = ({ inModal = false }: { inModal?: boolean }) => (
    <div
      className={cn(
        'rounded-lg border border-gray-700 bg-gray-900 overflow-hidden',
        !inModal && 'max-h-[600px]'
      )}
    >
      {showControls && (
        <div className="px-4 py-2 bg-gray-800 border-b border-gray-700 flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <FileCode className="w-4 h-4 text-blue-400" />
            <span className="text-sm text-gray-300 font-mono">
              {filename || 'Untitled'}
            </span>
            <span className="text-xs text-gray-500 bg-gray-700 px-2 py-1 rounded">
              {detectedLanguage}
            </span>
            <span className="text-xs text-gray-500">{lineCount} lines</span>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={handleDownload}
              className="p-1.5 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors"
              title="Download file"
            >
              <Download className="w-4 h-4" />
            </button>
            {!inModal && (
              <button
                onClick={() => setIsFullscreen(true)}
                className="p-1.5 text-gray-400 hover:text-white hover:bg-gray-700 rounded transition-colors"
                title="View fullscreen"
              >
                <Maximize2 className="w-4 h-4" />
              </button>
            )}
            <CopyButton content={code} className="text-gray-400 hover:text-white hover:bg-gray-700" />
          </div>
        </div>
      )}

      <div
        className={cn('overflow-auto', inModal ? 'max-h-[80vh]' : `max-h-[${maxHeight}px]`)}
      >
        <table className="w-full text-sm font-mono">
          <tbody>
            {lines.map((line, idx) => (
              <tr key={idx} className="hover:bg-gray-800/50">
                <td className="px-4 py-0.5 text-right text-gray-500 select-none w-12 align-top">
                  {idx + 1}
                </td>
                <td className="px-4 py-0.5 whitespace-pre text-gray-300">
                  <span dangerouslySetInnerHTML={{ __html: highlightCode(line) }} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  return (
    <>
      <CodeDisplay />

      {isFullscreen && (
        <div
          className="fixed inset-0 z-50 bg-black bg-opacity-90 flex items-center justify-center p-4"
          onClick={() => setIsFullscreen(false)}
        >
          <div
            className="relative max-w-[90vw] w-full max-h-[90vh]"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={() => setIsFullscreen(false)}
              className="absolute -top-10 right-0 p-2 text-white hover:text-gray-300 transition-colors"
              title="Close"
            >
              <X className="w-6 h-6" />
            </button>
            <CodeDisplay inModal />
          </div>
        </div>
      )}
    </>
  );
};
