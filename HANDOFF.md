# Phase 5: Advanced Features - Implementation Handoff

**Branch:** `phase-5-advanced-features`
**Epic:** `brandon-fryslie_claude-code-proxy-wa1` (request-comparison), `brandon-fryslie_claude-code-proxy-jaq` (conversation-threads), `brandon-fryslie_claude-code-proxy-5jo` (data-management)

---

## Executive Summary

You are implementing the **Advanced Features** for the new dashboard. These are power-user features that enable deep analysis and efficient data management.

**Prerequisites:** Phases 2-4 should be complete. You'll build on top of message rendering, tool display, and chart components.

Your deliverables:
1. **Request Comparison** - Side-by-side diff view for comparing two requests
2. **Conversation Threads** - Full message thread display with user/assistant bubbles
3. **Data Management** - Refresh, clear, settings persistence, auto-refresh

---

## Your Working Environment

### Directory Structure
```
/Users/bmf/code/brandon-fryslie_claude-code-proxy/
├── dashboard/              # NEW dashboard (your target)
│   ├── src/
│   │   ├── components/
│   │   │   ├── layout/
│   │   │   ├── ui/         # Phase 2-3 components
│   │   │   ├── charts/     # Phase 4 components
│   │   │   └── features/   # CREATE THIS - advanced features
│   │   ├── pages/
│   │   │   ├── Requests.tsx      # Add comparison mode
│   │   │   ├── Conversations.tsx # Enhance thread view
│   │   │   └── Settings.tsx      # Wire up settings
│   │   └── lib/
│   │       ├── types.ts
│   │       ├── api.ts
│   │       └── storage.ts  # CREATE THIS - localStorage utils
├── web/                    # OLD dashboard (reference)
│   └── app/
│       ├── routes/_index.tsx         # Compare mode logic
│       └── components/
│           ├── ConversationThread.tsx
│           └── MessageFlow.tsx
└── proxy/                  # Go backend
```

### API Endpoints You'll Use

```typescript
// Request Management
DELETE /api/requests           // Clear all requests
GET /api/requests/latest-date  // Get most recent request date

// Conversations
GET /api/conversations         // List conversations
GET /api/conversations/{id}    // Get single conversation with messages
GET /api/conversations/project // Get all conversations for a project
```

---

## Topic 1: Request Comparison

### What You're Building

A mode that allows users to select two requests and view them side-by-side, comparing:
- Request parameters (model, tokens, timing)
- Message content (what was sent)
- Response content (what Claude returned)
- Tool usage differences

### User Flow

1. User clicks "Compare" button in header → enters compare mode
2. Checkboxes appear next to each request in the list
3. User selects up to 2 requests
4. Banner shows selected count with "Compare" button
5. Click "Compare" → opens comparison modal
6. Click "Cancel" or Escape → exits compare mode

### Compare Mode State

```tsx
// In Requests.tsx or a parent component
interface CompareState {
  enabled: boolean;
  selectedIds: string[];  // Max 2
}

const [compareState, setCompareState] = useState<CompareState>({
  enabled: false,
  selectedIds: [],
});

const toggleCompareMode = () => {
  setCompareState(prev => ({
    enabled: !prev.enabled,
    selectedIds: [],
  }));
};

const toggleRequestSelection = (id: string) => {
  setCompareState(prev => {
    if (prev.selectedIds.includes(id)) {
      return {
        ...prev,
        selectedIds: prev.selectedIds.filter(x => x !== id),
      };
    }
    // Max 2 selected
    if (prev.selectedIds.length >= 2) {
      return {
        ...prev,
        selectedIds: [prev.selectedIds[1], id],  // Remove oldest, add new
      };
    }
    return {
      ...prev,
      selectedIds: [...prev.selectedIds, id],
    };
  });
};
```

### CompareModeBanner.tsx

```tsx
// dashboard/src/components/features/CompareModeBanner.tsx
import { type FC } from 'react';
import { GitCompare, X } from 'lucide-react';

interface CompareModeBannerProps {
  selectedCount: number;
  onCompare: () => void;
  onCancel: () => void;
}

export const CompareModeBanner: FC<CompareModeBannerProps> = ({
  selectedCount,
  onCompare,
  onCancel,
}) => {
  return (
    <div className="sticky top-0 z-40 bg-indigo-600 text-white px-4 py-2 flex items-center justify-between shadow-lg">
      <div className="flex items-center gap-3">
        <GitCompare className="w-5 h-5" />
        <span className="font-medium">Compare Mode</span>
        <span className="px-2 py-0.5 bg-white/20 rounded text-sm">
          {selectedCount}/2 selected
        </span>
      </div>

      <div className="flex items-center gap-2">
        <button
          onClick={onCompare}
          disabled={selectedCount !== 2}
          className="px-4 py-1.5 bg-white text-indigo-600 font-medium rounded-lg disabled:opacity-50 disabled:cursor-not-allowed hover:bg-indigo-50 transition-colors"
        >
          Compare Selected
        </button>
        <button
          onClick={onCancel}
          className="p-1.5 hover:bg-white/20 rounded-lg transition-colors"
          title="Exit compare mode"
        >
          <X className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
};
```

### RequestCompareModal.tsx

```tsx
// dashboard/src/components/features/RequestCompareModal.tsx
import { type FC, useMemo } from 'react';
import { X, ArrowRight, Clock, Zap, Hash } from 'lucide-react';
import { MessageContent } from '../ui/MessageContent';
import { formatTokens, formatDuration } from '@/lib/chartUtils';

interface RequestLog {
  request_id: string;
  timestamp: string;
  model: string;
  routed_model?: string;
  provider: string;
  body: {
    messages?: AnthropicMessage[];
    max_tokens?: number;
    temperature?: number;
    system?: string | SystemMessage[];
  };
  response?: {
    status_code: number;
    response_time: number;
    body?: {
      content?: ContentBlock[];
      usage?: {
        input_tokens: number;
        output_tokens: number;
      };
    };
  };
}

interface RequestCompareModalProps {
  request1: RequestLog;
  request2: RequestLog;
  onClose: () => void;
}

export const RequestCompareModal: FC<RequestCompareModalProps> = ({
  request1,
  request2,
  onClose,
}) => {
  // Calculate differences
  const diffs = useMemo(() => {
    const r1 = request1;
    const r2 = request2;

    return {
      model: r1.model !== r2.model,
      provider: r1.provider !== r2.provider,
      maxTokens: r1.body.max_tokens !== r2.body.max_tokens,
      temperature: r1.body.temperature !== r2.body.temperature,
      inputTokens: r1.response?.body?.usage?.input_tokens !== r2.response?.body?.usage?.input_tokens,
      outputTokens: r1.response?.body?.usage?.output_tokens !== r2.response?.body?.usage?.output_tokens,
      responseTime: Math.abs(
        (r1.response?.response_time || 0) - (r2.response?.response_time || 0)
      ) > 1000,  // > 1s difference
    };
  }, [request1, request2]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      {/* Backdrop */}
      <div
        className="absolute inset-0 bg-black/60"
        onClick={onClose}
      />

      {/* Modal */}
      <div className="relative bg-white rounded-2xl shadow-2xl max-w-[95vw] max-h-[90vh] w-full mx-4 overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b bg-gray-50">
          <h2 className="text-lg font-semibold text-gray-900">Compare Requests</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-200 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-500" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-auto">
          {/* Stats Comparison */}
          <div className="grid grid-cols-2 gap-4 p-4 bg-gray-50 border-b">
            <StatsCard request={request1} label="Request 1" />
            <StatsCard request={request2} label="Request 2" />
          </div>

          {/* Side-by-side Content */}
          <div className="grid grid-cols-2 divide-x">
            {/* Request 1 */}
            <div className="p-4 space-y-4">
              <SectionHeader title="Messages" requestId={request1.request_id} />
              <RequestContent request={request1} />
            </div>

            {/* Request 2 */}
            <div className="p-4 space-y-4">
              <SectionHeader title="Messages" requestId={request2.request_id} />
              <RequestContent request={request2} />
            </div>
          </div>

          {/* Response Comparison */}
          <div className="grid grid-cols-2 divide-x border-t">
            <div className="p-4 space-y-4">
              <SectionHeader title="Response" />
              <ResponseContent response={request1.response} />
            </div>
            <div className="p-4 space-y-4">
              <SectionHeader title="Response" />
              <ResponseContent response={request2.response} />
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

// Stats card for quick comparison
const StatsCard: FC<{ request: RequestLog; label: string }> = ({ request, label }) => {
  const usage = request.response?.body?.usage;

  return (
    <div className="bg-white rounded-lg p-4 border">
      <div className="text-xs text-gray-500 mb-2">{label}</div>
      <div className="grid grid-cols-4 gap-4 text-sm">
        <div>
          <div className="text-gray-500 text-xs">Model</div>
          <div className="font-medium truncate" title={request.model}>
            {getModelShortName(request.model)}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Response Time</div>
          <div className="font-medium flex items-center gap-1">
            <Clock className="w-3 h-3 text-gray-400" />
            {formatDuration(request.response?.response_time || 0)}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Input Tokens</div>
          <div className="font-medium flex items-center gap-1">
            <Hash className="w-3 h-3 text-gray-400" />
            {formatTokens(usage?.input_tokens || 0)}
          </div>
        </div>
        <div>
          <div className="text-gray-500 text-xs">Output Tokens</div>
          <div className="font-medium flex items-center gap-1">
            <Zap className="w-3 h-3 text-gray-400" />
            {formatTokens(usage?.output_tokens || 0)}
          </div>
        </div>
      </div>
    </div>
  );
};

const SectionHeader: FC<{ title: string; requestId?: string }> = ({ title, requestId }) => (
  <div className="flex items-center justify-between">
    <h3 className="font-medium text-gray-900">{title}</h3>
    {requestId && (
      <span className="text-xs font-mono text-gray-400">{requestId.slice(-8)}</span>
    )}
  </div>
);

const RequestContent: FC<{ request: RequestLog }> = ({ request }) => {
  const messages = request.body.messages || [];

  return (
    <div className="space-y-3 max-h-64 overflow-y-auto">
      {messages.map((msg, i) => (
        <div
          key={i}
          className={cn(
            "p-3 rounded-lg border",
            msg.role === 'user' ? 'bg-blue-50 border-blue-200' : 'bg-gray-50 border-gray-200'
          )}
        >
          <div className="text-xs font-medium text-gray-500 mb-1">
            {msg.role}
          </div>
          <MessageContent content={msg.content} />
        </div>
      ))}
    </div>
  );
};

const ResponseContent: FC<{ response?: RequestLog['response'] }> = ({ response }) => {
  if (!response?.body?.content) {
    return <div className="text-gray-400 italic">No response</div>;
  }

  return (
    <div className="max-h-64 overflow-y-auto">
      <MessageContent content={response.body.content} />
    </div>
  );
};

function getModelShortName(model: string): string {
  if (model.includes('opus')) return 'Opus';
  if (model.includes('sonnet')) return 'Sonnet';
  if (model.includes('haiku')) return 'Haiku';
  if (model.includes('gpt-4o')) return 'GPT-4o';
  if (model.includes('gpt-4')) return 'GPT-4';
  return model.split('-').slice(0, 2).join('-');
}

function cn(...classes: (string | boolean | undefined)[]): string {
  return classes.filter(Boolean).join(' ');
}
```

---

## Topic 2: Conversation Threads

### What You're Building

A full conversation thread view showing all messages between user and Claude, with:
- User/Assistant message bubbles (different colors)
- Tool calls displayed inline with their results
- System messages (collapsible)
- Timestamps for each message
- Connection lines between related messages

### API Response Structure

```typescript
interface ConversationDetail {
  sessionId: string;
  projectPath: string;
  projectName: string;
  messages: ConversationMessage[];
  startTime: string;
  endTime: string;
  messageCount: number;
}

interface ConversationMessage {
  parentUUID?: string;
  isSidechain: boolean;
  userType: string;
  cwd: string;
  sessionId: string;
  version: string;
  type: string;         // "user" | "assistant"
  message: unknown;     // Raw message content
  uuid: string;
  timestamp: string;
}
```

### ConversationThread.tsx

```tsx
// dashboard/src/components/features/ConversationThread.tsx
import { type FC, useMemo, useState } from 'react';
import { User, Bot, ChevronDown, Clock, GitBranch } from 'lucide-react';
import { MessageContent } from '../ui/MessageContent';
import { cn } from '@/lib/utils';

interface ConversationMessage {
  parentUUID?: string;
  isSidechain: boolean;
  type: string;
  message: unknown;
  uuid: string;
  timestamp: string;
  cwd?: string;
}

interface ConversationThreadProps {
  messages: ConversationMessage[];
  startTime: string;
  endTime: string;
}

export const ConversationThread: FC<ConversationThreadProps> = ({
  messages,
  startTime,
  endTime,
}) => {
  // Build message tree (handle parent/child relationships)
  const messageTree = useMemo(() => {
    return buildMessageTree(messages);
  }, [messages]);

  // Count by role
  const stats = useMemo(() => {
    const user = messages.filter(m => m.type === 'user').length;
    const assistant = messages.filter(m => m.type === 'assistant').length;
    return { user, assistant, total: messages.length };
  }, [messages]);

  return (
    <div className="flex flex-col h-full">
      {/* Header with stats */}
      <div className="px-4 py-3 bg-gray-50 border-b flex items-center justify-between">
        <div className="flex items-center gap-4 text-sm text-gray-600">
          <span className="flex items-center gap-1">
            <Clock className="w-4 h-4 text-gray-400" />
            {formatTimeRange(startTime, endTime)}
          </span>
          <span>{stats.total} messages</span>
        </div>
        <div className="flex items-center gap-3 text-sm">
          <span className="flex items-center gap-1 text-blue-600">
            <User className="w-4 h-4" />
            {stats.user}
          </span>
          <span className="flex items-center gap-1 text-green-600">
            <Bot className="w-4 h-4" />
            {stats.assistant}
          </span>
        </div>
      </div>

      {/* Message list */}
      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messageTree.map((msg, index) => (
          <MessageBubble
            key={msg.uuid}
            message={msg}
            isFirst={index === 0}
            isLast={index === messageTree.length - 1}
          />
        ))}
      </div>
    </div>
  );
};

// Individual message bubble
interface MessageBubbleProps {
  message: ConversationMessage & { children?: ConversationMessage[] };
  isFirst: boolean;
  isLast: boolean;
  depth?: number;
}

const MessageBubble: FC<MessageBubbleProps> = ({
  message,
  isFirst,
  isLast,
  depth = 0,
}) => {
  const [expanded, setExpanded] = useState(true);
  const isUser = message.type === 'user';
  const isSystem = message.type === 'system';

  // Extract content from message
  const content = useMemo(() => {
    return extractContent(message.message);
  }, [message.message]);

  // Role-based styling
  const bubbleStyles = {
    user: 'bg-blue-50 border-blue-200 ml-8',
    assistant: 'bg-gray-50 border-gray-200 mr-8',
    system: 'bg-amber-50 border-amber-200 mx-4',
  };

  const avatarStyles = {
    user: 'bg-blue-100 text-blue-600',
    assistant: 'bg-green-100 text-green-600',
    system: 'bg-amber-100 text-amber-600',
  };

  const roleType = isUser ? 'user' : isSystem ? 'system' : 'assistant';

  return (
    <div className={cn("relative", depth > 0 && "ml-8 border-l-2 border-gray-200 pl-4")}>
      {/* Sidechain indicator */}
      {message.isSidechain && (
        <div className="absolute -left-3 top-4 bg-purple-100 text-purple-600 p-1 rounded-full">
          <GitBranch className="w-3 h-3" />
        </div>
      )}

      <div className={cn(
        "rounded-xl border p-4",
        bubbleStyles[roleType]
      )}>
        {/* Header */}
        <div className="flex items-center justify-between mb-2">
          <div className="flex items-center gap-2">
            <div className={cn(
              "w-7 h-7 rounded-full flex items-center justify-center",
              avatarStyles[roleType]
            )}>
              {isUser ? <User className="w-4 h-4" /> : <Bot className="w-4 h-4" />}
            </div>
            <span className="font-medium text-sm text-gray-700 capitalize">
              {message.type}
            </span>
            <span className="text-xs text-gray-400">
              {formatTimestamp(message.timestamp)}
            </span>
          </div>

          {/* Collapse button for long content */}
          {content.length > 500 && (
            <button
              onClick={() => setExpanded(!expanded)}
              className="p-1 hover:bg-white/50 rounded"
            >
              <ChevronDown className={cn(
                "w-4 h-4 text-gray-400 transition-transform",
                !expanded && "-rotate-90"
              )} />
            </button>
          )}
        </div>

        {/* Content */}
        {expanded && (
          <div className="prose prose-sm max-w-none">
            <MessageContent content={content} showSystemReminders={isSystem} />
          </div>
        )}

        {!expanded && (
          <div className="text-sm text-gray-500 italic">
            {content.length > 100 ? content.slice(0, 100) + '...' : content}
          </div>
        )}
      </div>

      {/* Child messages (for threaded replies) */}
      {message.children?.map((child, i) => (
        <MessageBubble
          key={child.uuid}
          message={child}
          isFirst={false}
          isLast={i === (message.children?.length || 0) - 1}
          depth={depth + 1}
        />
      ))}
    </div>
  );
};

// Build tree from flat message list
function buildMessageTree(messages: ConversationMessage[]): (ConversationMessage & { children?: ConversationMessage[] })[] {
  const messageMap = new Map<string, ConversationMessage & { children?: ConversationMessage[] }>();
  const roots: (ConversationMessage & { children?: ConversationMessage[] })[] = [];

  // First pass: create map
  messages.forEach(msg => {
    messageMap.set(msg.uuid, { ...msg, children: [] });
  });

  // Second pass: build tree
  messages.forEach(msg => {
    const node = messageMap.get(msg.uuid)!;
    if (msg.parentUUID && messageMap.has(msg.parentUUID)) {
      const parent = messageMap.get(msg.parentUUID)!;
      parent.children = parent.children || [];
      parent.children.push(node);
    } else {
      roots.push(node);
    }
  });

  // Sort by timestamp
  roots.sort((a, b) => new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime());

  return roots;
}

// Extract content from various message formats
function extractContent(message: unknown): string | ContentBlock[] {
  if (!message) return '';

  // Handle different message structures
  if (typeof message === 'string') return message;

  if (typeof message === 'object') {
    const msg = message as Record<string, unknown>;

    // Direct content field
    if (msg.content) {
      if (typeof msg.content === 'string') return msg.content;
      if (Array.isArray(msg.content)) return msg.content;
    }

    // Wrapped in message object
    if (msg.message && typeof msg.message === 'object') {
      return extractContent(msg.message);
    }

    // Text field
    if (msg.text && typeof msg.text === 'string') return msg.text;
  }

  // Fallback: stringify
  return JSON.stringify(message, null, 2);
}

function formatTimestamp(timestamp: string): string {
  const date = new Date(timestamp);
  return date.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
    hour12: true,
  });
}

function formatTimeRange(start: string, end: string): string {
  const startDate = new Date(start);
  const endDate = new Date(end);
  const durationMs = endDate.getTime() - startDate.getTime();
  const durationMins = Math.round(durationMs / 60000);

  if (durationMins < 60) {
    return `${durationMins}m`;
  }
  return `${Math.round(durationMins / 60)}h ${durationMins % 60}m`;
}
```

### ConversationList.tsx

```tsx
// dashboard/src/components/features/ConversationList.tsx
import { type FC } from 'react';
import { MessageSquare, Clock, FolderOpen } from 'lucide-react';
import { cn } from '@/lib/utils';

interface Conversation {
  id: string;
  requestCount: number;
  startTime: string;
  lastActivity: string;
  duration: number;
  firstMessage: string;
  projectName: string;
}

interface ConversationListProps {
  conversations: Conversation[];
  selectedId?: string;
  onSelect: (id: string) => void;
  loading?: boolean;
}

export const ConversationList: FC<ConversationListProps> = ({
  conversations,
  selectedId,
  onSelect,
  loading,
}) => {
  if (loading) {
    return (
      <div className="p-4 space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="animate-pulse">
            <div className="h-20 bg-gray-100 rounded-lg" />
          </div>
        ))}
      </div>
    );
  }

  if (conversations.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-gray-400">
        <MessageSquare className="w-12 h-12 mb-3 opacity-50" />
        <p>No conversations found</p>
      </div>
    );
  }

  return (
    <div className="divide-y">
      {conversations.map(conv => (
        <ConversationItem
          key={conv.id}
          conversation={conv}
          isSelected={conv.id === selectedId}
          onClick={() => onSelect(conv.id)}
        />
      ))}
    </div>
  );
};

const ConversationItem: FC<{
  conversation: Conversation;
  isSelected: boolean;
  onClick: () => void;
}> = ({ conversation, isSelected, onClick }) => {
  return (
    <button
      onClick={onClick}
      className={cn(
        "w-full text-left p-4 hover:bg-gray-50 transition-colors",
        isSelected && "bg-blue-50 border-l-4 border-blue-500"
      )}
    >
      {/* Project name */}
      <div className="flex items-center gap-2 text-sm text-gray-500 mb-1">
        <FolderOpen className="w-3.5 h-3.5" />
        <span className="truncate">{conversation.projectName}</span>
      </div>

      {/* First message preview */}
      <div className="text-sm text-gray-900 line-clamp-2 mb-2">
        {conversation.firstMessage || 'No message preview'}
      </div>

      {/* Stats */}
      <div className="flex items-center gap-3 text-xs text-gray-400">
        <span className="flex items-center gap-1">
          <MessageSquare className="w-3 h-3" />
          {conversation.requestCount} messages
        </span>
        <span className="flex items-center gap-1">
          <Clock className="w-3 h-3" />
          {formatDuration(conversation.duration)}
        </span>
        <span>{formatRelativeTime(conversation.lastActivity)}</span>
      </div>
    </button>
  );
};

function formatDuration(ms: number): string {
  const mins = Math.round(ms / 60000);
  if (mins < 60) return `${mins}m`;
  const hours = Math.floor(mins / 60);
  return `${hours}h ${mins % 60}m`;
}

function formatRelativeTime(timestamp: string): string {
  const now = Date.now();
  const then = new Date(timestamp).getTime();
  const diffMins = Math.round((now - then) / 60000);

  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`;
  return `${Math.floor(diffMins / 1440)}d ago`;
}
```

---

## Topic 3: Data Management

### What You're Building

1. **Refresh Button** - Manual data refresh
2. **Clear All Data** - Delete request history with confirmation
3. **Settings Persistence** - Save preferences to localStorage
4. **Auto-Refresh** - Optional automatic data polling

### Local Storage Utilities

```tsx
// dashboard/src/lib/storage.ts

const STORAGE_PREFIX = 'claude-proxy-dashboard:';

export interface DashboardSettings {
  autoRefreshEnabled: boolean;
  autoRefreshInterval: number;  // seconds
  defaultModelFilter: string;
  darkMode: boolean;
  notifyOnError: boolean;
  notifyOnHighLatency: boolean;
  highLatencyThreshold: number;  // ms
  dataRetentionDays: number;
}

const DEFAULT_SETTINGS: DashboardSettings = {
  autoRefreshEnabled: false,
  autoRefreshInterval: 30,
  defaultModelFilter: 'all',
  darkMode: false,
  notifyOnError: true,
  notifyOnHighLatency: false,
  highLatencyThreshold: 5000,
  dataRetentionDays: 30,
};

export function getSettings(): DashboardSettings {
  try {
    const stored = localStorage.getItem(STORAGE_PREFIX + 'settings');
    if (stored) {
      return { ...DEFAULT_SETTINGS, ...JSON.parse(stored) };
    }
  } catch (e) {
    console.error('Failed to load settings:', e);
  }
  return DEFAULT_SETTINGS;
}

export function saveSettings(settings: Partial<DashboardSettings>): void {
  try {
    const current = getSettings();
    const updated = { ...current, ...settings };
    localStorage.setItem(STORAGE_PREFIX + 'settings', JSON.stringify(updated));
  } catch (e) {
    console.error('Failed to save settings:', e);
  }
}

export function getLastSelectedDate(): string | null {
  return localStorage.getItem(STORAGE_PREFIX + 'selectedDate');
}

export function saveSelectedDate(date: string): void {
  localStorage.setItem(STORAGE_PREFIX + 'selectedDate', date);
}

export function clearAllStorage(): void {
  Object.keys(localStorage)
    .filter(key => key.startsWith(STORAGE_PREFIX))
    .forEach(key => localStorage.removeItem(key));
}
```

### Settings Hook

```tsx
// dashboard/src/lib/hooks/useSettings.ts
import { useState, useEffect, useCallback } from 'react';
import { getSettings, saveSettings, type DashboardSettings } from '../storage';

export function useSettings() {
  const [settings, setSettingsState] = useState<DashboardSettings>(getSettings);

  // Sync with localStorage on mount
  useEffect(() => {
    setSettingsState(getSettings());
  }, []);

  const updateSettings = useCallback((updates: Partial<DashboardSettings>) => {
    setSettingsState(prev => {
      const updated = { ...prev, ...updates };
      saveSettings(updated);
      return updated;
    });
  }, []);

  const resetSettings = useCallback(() => {
    const defaults = getSettings();  // Will return defaults
    localStorage.removeItem('claude-proxy-dashboard:settings');
    setSettingsState(defaults);
  }, []);

  return {
    settings,
    updateSettings,
    resetSettings,
  };
}
```

### Auto-Refresh Hook

```tsx
// dashboard/src/lib/hooks/useAutoRefresh.ts
import { useEffect, useRef, useCallback } from 'react';
import { useSettings } from './useSettings';

interface UseAutoRefreshOptions {
  onRefresh: () => void | Promise<void>;
  enabled?: boolean;  // Override settings
}

export function useAutoRefresh({ onRefresh, enabled }: UseAutoRefreshOptions) {
  const { settings } = useSettings();
  const intervalRef = useRef<NodeJS.Timeout | null>(null);
  const lastRefreshRef = useRef<number>(Date.now());

  const isEnabled = enabled ?? settings.autoRefreshEnabled;
  const intervalMs = settings.autoRefreshInterval * 1000;

  const refresh = useCallback(async () => {
    lastRefreshRef.current = Date.now();
    await onRefresh();
  }, [onRefresh]);

  // Set up interval
  useEffect(() => {
    if (!isEnabled) {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
        intervalRef.current = null;
      }
      return;
    }

    intervalRef.current = setInterval(refresh, intervalMs);

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, [isEnabled, intervalMs, refresh]);

  // Manual refresh trigger
  const triggerRefresh = useCallback(() => {
    refresh();
  }, [refresh]);

  return {
    triggerRefresh,
    lastRefresh: lastRefreshRef.current,
    isAutoRefreshEnabled: isEnabled,
  };
}
```

### DataManagementBar.tsx

```tsx
// dashboard/src/components/features/DataManagementBar.tsx
import { type FC, useState } from 'react';
import { RefreshCw, Trash2, Settings, Loader2 } from 'lucide-react';
import { cn } from '@/lib/utils';

interface DataManagementBarProps {
  onRefresh: () => Promise<void>;
  onClearData: () => Promise<void>;
  isRefreshing?: boolean;
  lastRefresh?: Date;
}

export const DataManagementBar: FC<DataManagementBarProps> = ({
  onRefresh,
  onClearData,
  isRefreshing,
  lastRefresh,
}) => {
  const [showClearConfirm, setShowClearConfirm] = useState(false);
  const [isClearing, setIsClearing] = useState(false);

  const handleClear = async () => {
    setIsClearing(true);
    try {
      await onClearData();
      setShowClearConfirm(false);
    } finally {
      setIsClearing(false);
    }
  };

  return (
    <div className="flex items-center gap-2">
      {/* Refresh button */}
      <button
        onClick={onRefresh}
        disabled={isRefreshing}
        className={cn(
          "flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg transition-colors",
          "bg-gray-100 hover:bg-gray-200 text-gray-700",
          isRefreshing && "opacity-50 cursor-not-allowed"
        )}
        title={lastRefresh ? `Last refresh: ${lastRefresh.toLocaleTimeString()}` : 'Refresh data'}
      >
        <RefreshCw className={cn("w-4 h-4", isRefreshing && "animate-spin")} />
        Refresh
      </button>

      {/* Clear data button */}
      <div className="relative">
        <button
          onClick={() => setShowClearConfirm(true)}
          className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-red-50 hover:bg-red-100 text-red-600 transition-colors"
        >
          <Trash2 className="w-4 h-4" />
          Clear Data
        </button>

        {/* Confirmation popover */}
        {showClearConfirm && (
          <>
            <div
              className="fixed inset-0 z-40"
              onClick={() => setShowClearConfirm(false)}
            />
            <div className="absolute top-full mt-2 right-0 z-50 bg-white rounded-lg shadow-lg border p-4 w-64">
              <div className="text-sm font-medium text-gray-900 mb-2">
                Clear all request data?
              </div>
              <div className="text-xs text-gray-500 mb-4">
                This will permanently delete all logged requests. This action cannot be undone.
              </div>
              <div className="flex gap-2">
                <button
                  onClick={() => setShowClearConfirm(false)}
                  className="flex-1 px-3 py-1.5 text-sm rounded bg-gray-100 hover:bg-gray-200 transition-colors"
                >
                  Cancel
                </button>
                <button
                  onClick={handleClear}
                  disabled={isClearing}
                  className="flex-1 px-3 py-1.5 text-sm rounded bg-red-600 text-white hover:bg-red-700 transition-colors disabled:opacity-50"
                >
                  {isClearing ? (
                    <Loader2 className="w-4 h-4 mx-auto animate-spin" />
                  ) : (
                    'Delete All'
                  )}
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
};
```

### Enhanced Settings.tsx

```tsx
// dashboard/src/pages/Settings.tsx
import { type FC } from 'react';
import { Save, RotateCcw, Bell, Clock, Database } from 'lucide-react';
import { PageHeader, PageContent } from '@/components/layout';
import { useSettings } from '@/lib/hooks/useSettings';

export const Settings: FC = () => {
  const { settings, updateSettings, resetSettings } = useSettings();

  return (
    <>
      <PageHeader
        title="Settings"
        description="Configure dashboard preferences"
        actions={
          <button
            onClick={resetSettings}
            className="flex items-center gap-2 px-3 py-1.5 text-sm rounded-lg bg-gray-100 hover:bg-gray-200 text-gray-700 transition-colors"
          >
            <RotateCcw className="w-4 h-4" />
            Reset to Defaults
          </button>
        }
      />

      <PageContent>
        <div className="max-w-2xl space-y-8">
          {/* Auto-Refresh Section */}
          <SettingsSection
            icon={Clock}
            title="Auto-Refresh"
            description="Automatically refresh data at regular intervals"
          >
            <ToggleSetting
              label="Enable auto-refresh"
              checked={settings.autoRefreshEnabled}
              onChange={(checked) => updateSettings({ autoRefreshEnabled: checked })}
            />
            <SelectSetting
              label="Refresh interval"
              value={String(settings.autoRefreshInterval)}
              options={[
                { value: '15', label: '15 seconds' },
                { value: '30', label: '30 seconds' },
                { value: '60', label: '1 minute' },
                { value: '300', label: '5 minutes' },
              ]}
              onChange={(value) => updateSettings({ autoRefreshInterval: Number(value) })}
              disabled={!settings.autoRefreshEnabled}
            />
          </SettingsSection>

          {/* Notifications Section */}
          <SettingsSection
            icon={Bell}
            title="Notifications"
            description="Get alerted about important events"
          >
            <ToggleSetting
              label="Error notifications"
              description="Show notification when a request fails"
              checked={settings.notifyOnError}
              onChange={(checked) => updateSettings({ notifyOnError: checked })}
            />
            <ToggleSetting
              label="High latency warnings"
              description="Alert when response time exceeds threshold"
              checked={settings.notifyOnHighLatency}
              onChange={(checked) => updateSettings({ notifyOnHighLatency: checked })}
            />
            <NumberSetting
              label="Latency threshold"
              value={settings.highLatencyThreshold}
              onChange={(value) => updateSettings({ highLatencyThreshold: value })}
              suffix="ms"
              min={1000}
              max={30000}
              step={1000}
              disabled={!settings.notifyOnHighLatency}
            />
          </SettingsSection>

          {/* Data Management Section */}
          <SettingsSection
            icon={Database}
            title="Data Retention"
            description="Control how long data is stored"
          >
            <SelectSetting
              label="Keep request logs for"
              value={String(settings.dataRetentionDays)}
              options={[
                { value: '7', label: '7 days' },
                { value: '14', label: '14 days' },
                { value: '30', label: '30 days' },
                { value: '90', label: '90 days' },
                { value: '0', label: 'Forever' },
              ]}
              onChange={(value) => updateSettings({ dataRetentionDays: Number(value) })}
            />
            <div className="text-xs text-gray-400 mt-2">
              Note: Data retention is applied server-side. Changes take effect on next cleanup cycle.
            </div>
          </SettingsSection>
        </div>
      </PageContent>
    </>
  );
};

// Helper components
const SettingsSection: FC<{
  icon: FC<{ className?: string }>;
  title: string;
  description: string;
  children: React.ReactNode;
}> = ({ icon: Icon, title, description, children }) => (
  <div className="bg-white rounded-xl border p-6">
    <div className="flex items-start gap-3 mb-4">
      <div className="p-2 bg-gray-100 rounded-lg">
        <Icon className="w-5 h-5 text-gray-600" />
      </div>
      <div>
        <h3 className="font-medium text-gray-900">{title}</h3>
        <p className="text-sm text-gray-500">{description}</p>
      </div>
    </div>
    <div className="space-y-4 ml-12">{children}</div>
  </div>
);

const ToggleSetting: FC<{
  label: string;
  description?: string;
  checked: boolean;
  onChange: (checked: boolean) => void;
  disabled?: boolean;
}> = ({ label, description, checked, onChange, disabled }) => (
  <label className={cn("flex items-start gap-3 cursor-pointer", disabled && "opacity-50 cursor-not-allowed")}>
    <input
      type="checkbox"
      checked={checked}
      onChange={(e) => onChange(e.target.checked)}
      disabled={disabled}
      className="mt-1 w-4 h-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
    />
    <div>
      <div className="text-sm font-medium text-gray-700">{label}</div>
      {description && <div className="text-xs text-gray-400">{description}</div>}
    </div>
  </label>
);

const SelectSetting: FC<{
  label: string;
  value: string;
  options: { value: string; label: string }[];
  onChange: (value: string) => void;
  disabled?: boolean;
}> = ({ label, value, options, onChange, disabled }) => (
  <div className={cn("flex items-center justify-between", disabled && "opacity-50")}>
    <span className="text-sm text-gray-700">{label}</span>
    <select
      value={value}
      onChange={(e) => onChange(e.target.value)}
      disabled={disabled}
      className="px-3 py-1.5 text-sm border rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
    >
      {options.map(opt => (
        <option key={opt.value} value={opt.value}>{opt.label}</option>
      ))}
    </select>
  </div>
);

const NumberSetting: FC<{
  label: string;
  value: number;
  onChange: (value: number) => void;
  suffix?: string;
  min?: number;
  max?: number;
  step?: number;
  disabled?: boolean;
}> = ({ label, value, onChange, suffix, min, max, step, disabled }) => (
  <div className={cn("flex items-center justify-between", disabled && "opacity-50")}>
    <span className="text-sm text-gray-700">{label}</span>
    <div className="flex items-center gap-2">
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        min={min}
        max={max}
        step={step}
        disabled={disabled}
        className="w-24 px-3 py-1.5 text-sm border rounded-lg text-right focus:ring-2 focus:ring-blue-500"
      />
      {suffix && <span className="text-sm text-gray-400">{suffix}</span>}
    </div>
  </div>
);

function cn(...classes: (string | boolean | undefined)[]): string {
  return classes.filter(Boolean).join(' ');
}
```

---

## Testing Checklist

### Request Comparison
- [ ] Compare mode toggle shows banner
- [ ] Checkboxes appear on each request
- [ ] Can select up to 2 requests
- [ ] Selecting 3rd replaces oldest
- [ ] Compare button disabled until 2 selected
- [ ] Modal shows side-by-side comparison
- [ ] Stats differences highlighted
- [ ] Messages displayed correctly
- [ ] Responses displayed correctly
- [ ] Escape key closes modal
- [ ] Cancel button exits compare mode

### Conversation Threads
- [ ] Conversation list shows all conversations
- [ ] Selected conversation highlighted
- [ ] Message bubbles show correct colors (user/assistant)
- [ ] System messages collapsible
- [ ] Sidechain messages indicated
- [ ] Timestamps displayed
- [ ] Parent/child relationships shown
- [ ] Long messages truncated with expand
- [ ] Empty conversations handled

### Data Management
- [ ] Refresh button triggers data reload
- [ ] Refresh shows loading state
- [ ] Clear data shows confirmation
- [ ] Clear data actually deletes (API call)
- [ ] Settings persist to localStorage
- [ ] Settings load on page refresh
- [ ] Reset button restores defaults
- [ ] Auto-refresh works when enabled
- [ ] Auto-refresh stops when disabled

---

## Common Gotchas

1. **Compare Mode State**: Keep in parent component (Requests.tsx), not in list item
2. **Message Trees**: Handle orphaned messages (parentUUID points to non-existent message)
3. **localStorage Errors**: Wrap in try/catch (can fail in incognito mode)
4. **Auto-refresh Memory Leaks**: Clear intervals on unmount
5. **Modal Focus Trap**: Trap focus inside modal for accessibility
6. **Large Conversations**: Virtual scrolling may be needed for 100+ messages

---

## Reference Files

Study these in the old dashboard:
- `web/app/routes/_index.tsx` - Compare mode implementation
- `web/app/components/ConversationThread.tsx` - Thread display
- `web/app/components/MessageFlow.tsx` - Message bubbles

---

## Definition of Done

Phase 5 is complete when:

1. Request comparison works with side-by-side view
2. Conversation threads display with proper bubbles
3. Settings persist to localStorage
4. Refresh button works
5. Clear data works with confirmation
6. Auto-refresh works when enabled
7. All components handle edge cases gracefully
8. No TypeScript errors in strict mode
9. Commit history shows logical, atomic commits

---

**You're building the power features. Make advanced users happy.**
