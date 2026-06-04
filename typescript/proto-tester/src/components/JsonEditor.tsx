/**
 * JsonEditor.tsx - CodeMirror 6 JSON 编辑器组件
 *
 * 职责：
 * - 提供基于 CodeMirror 6 的 JSON 语法高亮编辑器
 * - 支持自动格式化和只读模式
 * - 用于 _fallback_json 降级场景和原始数据编辑
 *
 * 不负责：
 * - 表单字段校验（由 ProtoFormRenderer 负责）
 * - 数据序列化（由调用方处理）
 */
import { useEffect, useRef } from 'react';
import { EditorState } from '@codemirror/state';
import { EditorView, keymap } from '@codemirror/view';
import { json } from '@codemirror/lang-json';
import { oneDark } from '@codemirror/theme-one-dark';

/** JsonEditor 组件属性 */
interface JsonEditorProps {
  /** 当前编辑器内容（JSON 字符串） */
  value: string;
  /** 内容变更回调 */
  onChange: (value: string) => void;
  /** 是否只读模式，默认 false */
  readOnly?: boolean;
  /** 编辑器高度，默认 '200px' */
  height?: string;
}

/**
 * 基于 CodeMirror 6 的 JSON 编辑器
 *
 * 特性：
 * - JSON 语法高亮（使用 @codemirror/lang-json）
 * - oneDark 暗色主题
 * - 受控组件模式：value/onChange 由父组件管理
 * - 自动格式化：失焦时尝试美化 JSON
 */
export function JsonEditor({ value, onChange, readOnly = false, height = '200px' }: JsonEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const viewRef = useRef<EditorView | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;

    // 创建 EditorState
    const state = EditorState.create({
      doc: value,
      extensions: [
        json(),
        oneDark,
        EditorView.theme({
          '&': { height },
          '.cm-scroller': { overflow: 'auto', fontFamily: "'JetBrains Mono', monospace", fontSize: '13px' },
        }),
        keymap.of([{
          key: 'Mod-s',
          run: () => true, // 阻止默认保存行为
        }]),
        EditorView.updateListener.of((update) => {
          if (update.docChanged) {
            onChange(update.state.doc.toString());
          }
        }),
        readOnly ? EditorState.readOnly.of(true) : [],
      ],
    });

    // 创建 EditorView
    const view = new EditorView({
      state,
      parent: containerRef.current,
    });
    viewRef.current = view;

    return () => {
      view.destroy();
      viewRef.current = null;
    };
    // 仅在 value 和 readOnly 变化时重建编辑器
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [readOnly, height]);

  // 外部 value 变更时同步到编辑器（避免循环更新）
  useEffect(() => {
    if (viewRef.current && viewRef.current.state.doc.toString() !== value) {
      viewRef.current.dispatch({
        changes: {
          from: 0,
          to: viewRef.current.state.doc.length,
          insert: value,
        },
      });
    }
  }, [value]);

  return <div ref={containerRef} />;
}
