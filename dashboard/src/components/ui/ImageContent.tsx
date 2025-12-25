import { type FC } from 'react'

export interface ImageSource {
  type: 'base64'
  media_type: 'image/jpeg' | 'image/png' | 'image/gif' | 'image/webp'
  data: string
}

interface ImageContentProps {
  source: ImageSource
}

export const ImageContent: FC<ImageContentProps> = ({ source }) => {
  if (source.type !== 'base64') {
    return (
      <div className="text-sm text-red-500">
        Unsupported image source type: {source.type}
      </div>
    )
  }

  const dataUrl = `data:${source.media_type};base64,${source.data}`

  return (
    <div className="my-2">
      <img
        src={dataUrl}
        alt="Message attachment"
        className="max-w-full h-auto rounded-lg border border-gray-200"
      />
    </div>
  )
}
