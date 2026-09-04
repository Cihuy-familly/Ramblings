import { useState, useCallback } from 'react'
import MDEditor, { commands, type ICommand } from '@uiw/react-md-editor'
import { uploadImage } from '../api/client'
import { Image, Loader } from 'lucide-react'

interface MarkdownEditorProps {
  value: string
  onChange: (value: string) => void
  height?: number
}

export default function MarkdownEditor({ value, onChange, height = 500 }: MarkdownEditorProps) {
  const [uploading, setUploading] = useState(false)

  const handleImageUpload = useCallback(async () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = async () => {
      const file = input.files?.[0]
      if (!file) return

      setUploading(true)
      try {
        const url = await uploadImage(file)
        // Insert image markdown at cursor position
        const imageMarkdown = `\n![image](${url})\n`
        onChange(value + imageMarkdown)
      } catch (err) {
        console.error('Image upload failed:', err)
      } finally {
        setUploading(false)
      }
    }
    input.click()
  }, [value, onChange])

  const imageCommand: ICommand = {
    name: 'image-upload',
    keyCommand: 'image',
    buttonProps: { 'aria-label': 'Upload image', title: 'Upload image' },
    icon: (
      <span className="flex items-center justify-center">
        {uploading ? (
          <Loader className="w-4 h-4 animate-spin" />
        ) : (
          <Image className="w-4 h-4" />
        )}
      </span>
    ),
    execute: () => {
      handleImageUpload()
    },
  }

  return (
    <div data-color-mode="light">
      <MDEditor
        value={value}
        onChange={(val) => onChange(val || '')}
        height={height}
        preview="live"
        visibleDragbar={false}
        commands={[
          commands.bold,
          commands.italic,
          commands.strikethrough,
          commands.hr,
          commands.title,
          commands.divider,
          commands.link,
          commands.quote,
          commands.code,
          commands.codeBlock,
          commands.divider,
          imageCommand,
          commands.divider,
          commands.unorderedListCommand,
          commands.orderedListCommand,
          commands.checkedListCommand,
        ]}
      />
    </div>
  )
}