import { type FC } from 'react';
import { AlertCircle, CheckCircle } from 'lucide-react';
import { CodeViewer } from '../CodeViewer';
import { FileListContent } from './FileListContent';

interface ToolResultContentProps {
  content: string;
  isError?: boolean;
}

interface ContentAnalysis {
  type: 'code' | 'json' | 'file_list' | 'error' | 'table' | 'text';
  language?: string;
  metadata?: Record<string, unknown>;
}

export const ToolResultContent: FC<ToolResultContentProps> = ({ content, isError = false }) => {
  const analysis = analyzeContent(content);

  // Error content
  if (isError || analysis.type === 'error') {
    return (
      <div className="flex items-start gap-3 p-4 bg-red-50 border border-red-200 rounded-lg text-red-700">
        <AlertCircle className="w-5 h-5 flex-shrink-0 mt-0.5" />
        <div className="flex-1 min-w-0">
          <div className="font-medium mb-1">Error</div>
          <pre className="text-sm whitespace-pre-wrap font-mono">{content}</pre>
        </div>
      </div>
    );
  }

  // Success indicator for empty/short success messages
  if (content.trim() === '' || content.trim().toLowerCase() === 'success') {
    return (
      <div className="flex items-center gap-2 p-3 bg-green-50 border border-green-200 rounded-lg text-green-700">
        <CheckCircle className="w-4 h-4" />
        <span className="text-sm font-medium">Success</span>
      </div>
    );
  }

  // JSON content
  if (analysis.type === 'json') {
    try {
      const parsed = JSON.parse(content);
      const formatted = JSON.stringify(parsed, null, 2);
      return <CodeViewer code={formatted} language="json" maxHeight={300} />;
    } catch {
      // Fall through to text
    }
  }

  // File list
  if (analysis.type === 'file_list') {
    return <FileListContent content={content} />;
  }

  // Code content
  if (analysis.type === 'code') {
    return (
      <CodeViewer code={content} language={analysis.language || 'text'} maxHeight={400} />
    );
  }

  // Table format
  if (analysis.type === 'table') {
    return (
      <div className="overflow-x-auto">
        <pre className="text-xs font-mono bg-gray-50 p-3 rounded border border-gray-200 whitespace-pre">
          {content}
        </pre>
      </div>
    );
  }

  // Plain text (default)
  return (
    <div className="text-sm text-gray-700 whitespace-pre-wrap bg-gray-50 p-3 rounded border border-gray-200 max-h-96 overflow-y-auto">
      {content}
    </div>
  );
};

function analyzeContent(content: string): ContentAnalysis {
  const trimmed = content.trim();

  // Error detection
  if (
    trimmed.startsWith('Error:') ||
    trimmed.startsWith('error:') ||
    /^(ENOENT|EACCES|EPERM|EEXIST)/.test(trimmed) ||
    trimmed.includes('Permission denied') ||
    trimmed.includes('No such file or directory')
  ) {
    return { type: 'error' };
  }

  // JSON detection
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    try {
      JSON.parse(trimmed);
      return { type: 'json' };
    } catch {
      // Not valid JSON, continue checking
    }
  }

  // File list (from glob/find)
  if (isFileList(trimmed)) {
    return { type: 'file_list' };
  }

  // Code with line numbers (cat -n format)
  if (/^\s*\d+[→\t]/.test(trimmed)) {
    const language = detectLanguageFromContent(trimmed);
    return { type: 'code', language };
  }

  // Code by content patterns
  if (hasCodePatterns(trimmed)) {
    const language = detectLanguageFromContent(trimmed);
    return { type: 'code', language };
  }

  // Table format (from ls -la, ps, etc.)
  if (isTableFormat(trimmed)) {
    return { type: 'table' };
  }

  return { type: 'text' };
}

function isFileList(content: string): boolean {
  const lines = content.split('\n').filter((l) => l.trim());
  if (lines.length < 2) return false;

  // Check if most lines look like file paths
  const pathLikeLines = lines.filter(
    (line) =>
      /^\.?\//.test(line) || // Starts with / or ./
      /\.(ts|tsx|js|jsx|py|go|rs|md|json|yaml|yml|toml|txt|css|html)$/.test(line) ||
      /^[a-zA-Z0-9_-]+\//.test(line) // Looks like relative path
  );

  return pathLikeLines.length >= lines.length * 0.7;
}

function isTableFormat(content: string): boolean {
  const lines = content.split('\n').filter((l) => l.trim());
  if (lines.length < 2) return false;

  // Check for consistent column structure (multiple whitespace-separated columns)
  const columnCounts = lines.map((line) => line.trim().split(/\s{2,}/).length);

  const avgColumns = columnCounts.reduce((a, b) => a + b, 0) / columnCounts.length;
  return avgColumns >= 3 && columnCounts.every((c) => Math.abs(c - avgColumns) <= 2);
}

function hasCodePatterns(content: string): boolean {
  const patterns = [
    /^(import|from|export|const|let|var|function|class|def|func|package)\s/m,
    /[{}\[\]];?\s*$/m,
    /^\s*(if|for|while|return|throw|try|catch)\s*[\(\{]/m,
    /=>\s*[{\(]/,
    /\bexport\s+(default\s+)?/,
    /^#!/, // Shebang
  ];
  return patterns.some((p) => p.test(content));
}

function detectLanguageFromContent(content: string): string {
  // Python indicators
  if (/^(def|class|import|from)\s/.test(content) || /:\s*$/.test(content.split('\n')[0] || '')) {
    return 'python';
  }

  // Go indicators
  if (/^(func|package|import)\s/.test(content) || /\bfmt\./.test(content)) {
    return 'go';
  }

  // Rust indicators
  if (/^(fn|use|mod|impl|struct|enum)\s/.test(content) || /\bmut\s/.test(content)) {
    return 'rust';
  }

  // TypeScript/JavaScript (most common in this context)
  if (/^(const|let|var|function|class|interface|type)\s/.test(content)) {
    return 'typescript';
  }

  // Bash
  if (/^#!/.test(content) || /\$\{?\w+\}?/.test(content)) {
    return 'bash';
  }

  return 'text';
}
